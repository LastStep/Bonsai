package selfcheck

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// The dist checks (design/plan-5.md, note 5.7.1, 4) hold a folder of release files to what a release must be:
// GoReleaser's dist/ after a snapshot or a release build, or the files downloaded from a published release. The
// folder is BONSAI_CHECK_DIST; without it TestDist is skipped. CI's release-check job runs it on every push against a
// fresh snapshot; the release workflow runs it before it signs. Only the archives (*.tar.gz, *.zip) and
// checksums.txt are read; GoReleaser's other files there (its build folders, artifacts.json, metadata.json,
// config.yaml) are not release files and are left alone.
//
//   - the archives: exactly one per system and processor .goreleaser.yaml builds, named by its name template, all of
//     one version, which reads X.Y.Z with an optional -suffix (a base-v pack tag taken for the version fails here);
//   - each archive: exactly the program (bonsai, or bonsai.exe in a zip), install.sh (0755 in a tar), install.ps1,
//     README.md and LICENSE, all at its top, nothing else;
//   - each program: none of FaultMarkers in its bytes; its build settings (debug/buildinfo): no -tags,
//     CGO_ENABLED=0, -trimpath=true, vcs.modified=false and a vcs.revision, equal to BONSAI_CHECK_COMMIT when set;
//   - checksums.txt: one "<sha256>  <archive>" line per archive, each right, and nothing else;
//   - the installer from its archive: on Linux the linux archive's install.sh, on Windows the windows zip's
//     install.ps1, each to a scratch target with a scratch Bonsai home, its install.json read back by check's reader,
//     and the installed --version naming the version.
//
// Run it by hand on a local snapshot (from the module root):
//
//	go run github.com/goreleaser/goreleaser/v2@v2.18.3 release --snapshot --clean
//	BONSAI_CHECK_DIST="$PWD/dist" go test -count=1 -run TestDist -v ./internal/selfcheck/

// distFolder and distCommit are the environment variables TestDist reads.
const (
	distFolder = "BONSAI_CHECK_DIST"
	distCommit = "BONSAI_CHECK_COMMIT"
)

// releaseVersion is a release's version as the archive names hold it: no leading v, an optional suffix
// (1.0.0, 1.0.0-rc.1, or a snapshot's 0.4.3-SNAPSHOT-2f18d892).
var releaseVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`)

// archiveFiles are the files besides the program every archive holds at its top, each with the mode it must have in
// a tar archive (0 for any).
var archiveFiles = map[string]os.FileMode{"install.sh": 0o755, "install.ps1": 0, "README.md": 0, "LICENSE": 0}

// releaseConfig is what TestDist reads of .goreleaser.yaml, as text, the way internal/engine's name test does.
type releaseConfig struct {
	project, nameTemplate string
	goos, goarch          []string
}

func readReleaseConfig(t *testing.T) releaseConfig {
	t.Helper()
	text := read(t, ".goreleaser.yaml")
	section := func(name string) string {
		m := regexp.MustCompile(`(?ms)^` + name + `:\n(.*?)(^\S|\z)`).FindStringSubmatch(text)
		if m == nil {
			t.Fatalf(".goreleaser.yaml has no %s section", name)
		}
		return m[1]
	}
	field := func(in, key string) string {
		m := regexp.MustCompile(`(?m)^\s*` + key + `:\s*"?([^"\n]*)"?\s*$`).FindStringSubmatch(in)
		if m == nil {
			t.Fatalf(".goreleaser.yaml has no %s in:\n%s", key, in)
		}
		return m[1]
	}
	list := func(in, key string) []string {
		m := regexp.MustCompile(`(?m)^(\s*)` + key + `:\n((?:\s*- \S+\n)+)`).FindStringSubmatch(in)
		if m == nil {
			t.Fatalf(".goreleaser.yaml has no %s list in:\n%s", key, in)
		}
		var out []string
		for _, l := range strings.Split(strings.TrimSpace(m[2]), "\n") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "- ")))
		}
		return out
	}
	builds := section("builds")
	return releaseConfig{
		project:      field(text, "project_name"),
		nameTemplate: field(section("archives"), "name_template"),
		goos:         list(builds, "goos"),
		goarch:       list(builds, "goarch"),
	}
}

// archiveName is the archive .goreleaser.yaml names for a version, system and processor.
func (c releaseConfig) archiveName(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return strings.NewReplacer("{{ .ProjectName }}", c.project, "{{ .Version }}", version, "{{ .Os }}", goos,
		"{{ .Arch }}", goarch).Replace(c.nameTemplate) + ext
}

// versionOf reads the version out of an archive's name: the name template with one of the build's systems and
// processors around it. It returns "" for a name the template cannot give.
func (c releaseConfig) versionOf(name string) string {
	const marker = "\x00"
	for _, goos := range c.goos {
		for _, goarch := range c.goarch {
			pattern := c.archiveName(marker, goos, goarch)
			i := strings.Index(pattern, marker)
			prefix, suffix := pattern[:i], pattern[i+len(marker):]
			if len(name) > len(prefix)+len(suffix) && strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
				return name[len(prefix) : len(name)-len(suffix)]
			}
		}
	}
	return ""
}

// unpacked is one archive's files, written into a temporary folder.
type unpacked struct {
	dir   string
	modes map[string]os.FileMode // each entry's name, as the archive holds it, and its mode
}

// unpack writes a tar.gz's or a zip's regular files into dir and returns each entry's name and mode. An entry that
// is a folder, a link or anything but a regular file is reported as a problem, and not written.
func unpack(path, dir string) (unpacked, []string, error) {
	u := unpacked{dir: dir, modes: map[string]os.FileMode{}}
	var problems []string
	put := func(name string, mode os.FileMode, regular bool, r io.Reader) error {
		u.modes[name] = mode
		if !regular {
			problems = append(problems, fmt.Sprintf("%s is not a regular file (mode %v)", name, mode))
			return nil
		}
		if name == "" || strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
			problems = append(problems, fmt.Sprintf("%q is not a file at the archive's top", name))
			return nil
		}
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, r)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		return err
	}
	if strings.HasSuffix(path, ".zip") {
		z, err := zip.OpenReader(path)
		if err != nil {
			return u, nil, err
		}
		defer func() { _ = z.Close() }()
		for _, f := range z.File {
			r, err := f.Open()
			if err != nil {
				return u, nil, err
			}
			err = put(f.Name, f.Mode(), f.Mode().IsRegular(), r)
			_ = r.Close()
			if err != nil {
				return u, nil, err
			}
		}
		return u, problems, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return u, nil, err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return u, nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return u, nil, err
		}
		if err := put(h.Name, h.FileInfo().Mode(), h.Typeflag == tar.TypeReg, tr); err != nil {
			return u, nil, err
		}
	}
	return u, problems, nil
}

func fileSum(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func TestDist(t *testing.T) {
	dir := os.Getenv(distFolder)
	if dir == "" {
		t.Skip(distFolder + " names no folder of release files (GoReleaser's dist/ or a release's downloads): " +
			"the dist checks run in CI's release-check job, and by hand as dist_test.go's header says")
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := readReleaseConfig(t)

	// The archives: one per system and processor, all of one version.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("%s: %v", distFolder, err)
	}
	var names []string
	for _, e := range entries {
		if n := e.Name(); !e.IsDir() && (strings.HasSuffix(n, ".tar.gz") || strings.HasSuffix(n, ".zip")) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		t.Fatalf("%s (%s) holds no archive", distFolder, filepath.ToSlash(dir))
	}
	version := cfg.versionOf(names[0])
	if !releaseVersion.MatchString(version) {
		t.Fatalf("%s: the version %q is not X.Y.Z with an optional -suffix (was a tag that is not Bonsai's taken for it?)",
			names[0], version)
	}
	want := map[string]bool{}
	for _, goos := range cfg.goos {
		for _, goarch := range cfg.goarch {
			want[cfg.archiveName(version, goos, goarch)] = true
		}
	}
	for _, n := range names {
		if !want[n] {
			t.Errorf("%s is not an archive .goreleaser.yaml makes for version %s", n, version)
		}
	}
	for _, n := range sortedKeys(want) {
		if !contains(names, n) {
			t.Errorf("%s is missing", n)
		}
	}

	// checksums.txt: one right line per archive, nothing else.
	t.Run("checksums", func(t *testing.T) {
		raw, err := os.ReadFile(filepath.Join(dir, "checksums.txt"))
		if err != nil {
			t.Fatal(err)
		}
		line := regexp.MustCompile(`^([0-9a-f]{64})  (\S+)$`)
		seen := map[string]bool{}
		for i, l := range strings.Split(strings.TrimSuffix(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n"), "\n") {
			m := line.FindStringSubmatch(l)
			if m == nil {
				t.Errorf("checksums.txt:%d: %q is not \"<sha256>  <file>\"", i+1, l)
				continue
			}
			if seen[m[2]] {
				t.Errorf("checksums.txt:%d: %s is listed twice", i+1, m[2])
			}
			seen[m[2]] = true
			if !contains(names, m[2]) {
				t.Errorf("checksums.txt:%d: %s is not an archive here", i+1, m[2])
				continue
			}
			if got := fileSum(t, filepath.Join(dir, m[2])); got != m[1] {
				t.Errorf("checksums.txt:%d: %s's SHA-256 is %s, the line says %s", i+1, m[2], got, m[1])
			}
		}
		for _, n := range names {
			if !seen[n] {
				t.Errorf("checksums.txt has no line for %s", n)
			}
		}
	})

	// Each archive's files, and its program.
	// Unpacked under the test's own folder, not each subtest's, so the installer's subtest finds them.
	unpackedAt, base := map[string]unpacked{}, t.TempDir()
	for i, n := range names {
		t.Run("archive/"+n, func(t *testing.T) {
			into := filepath.Join(base, fmt.Sprint(i))
			if err := os.Mkdir(into, 0o700); err != nil {
				t.Fatal(err)
			}
			u, problems, err := unpack(filepath.Join(dir, n), into)
			if err != nil {
				t.Fatalf("unpacking %s: %v", n, err)
			}
			for _, p := range problems {
				t.Error(p)
			}
			program := "bonsai"
			if strings.HasSuffix(n, ".zip") {
				program = "bonsai.exe"
			}
			files := map[string]os.FileMode{program: 0}
			for f, m := range archiveFiles {
				files[f] = m
			}
			for _, f := range sortedKeys(files) {
				m, ok := u.modes[f]
				if !ok {
					t.Errorf("%s holds no %s", n, f)
					continue
				}
				if files[f] != 0 && strings.HasSuffix(n, ".tar.gz") && m.Perm() != files[f] {
					t.Errorf("%s's %s has mode %v, want %v", n, f, m.Perm(), files[f])
				}
			}
			for _, f := range sortedKeys(u.modes) {
				if _, ok := files[f]; !ok {
					t.Errorf("%s holds %s, which a release archive does not", n, f)
				}
			}
			if _, ok := u.modes[program]; ok {
				checkProgram(t, n, filepath.Join(u.dir, program))
			}
			unpackedAt[n] = u
		})
	}

	// The installer from its archive, on this system.
	t.Run("installer", func(t *testing.T) {
		archive := cfg.archiveName(version, runtime.GOOS, runtime.GOARCH)
		u, ok := unpackedAt[archive]
		switch {
		case runtime.GOOS != "linux" && runtime.GOOS != "windows":
			t.Skip("the installers are Linux's and Windows'; " + runtime.GOOS + " has none")
		case !ok:
			t.Fatalf("no unpacked %s to install from", archive)
		}
		installFrom(t, u.dir, version)
	})
}

// checkProgram holds one program to the release's build: no fault code, and the build settings a release has.
func checkProgram(t *testing.T, archive, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range FaultMarkers() {
		if bytes.Contains(b, []byte(m)) {
			t.Errorf("%s's program holds %q: it was built with the test fault switch", archive, m)
		}
	}
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		t.Fatalf("%s's program: no build settings: %v", archive, err)
	}
	settings := map[string]string{}
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	if tags, ok := settings["-tags"]; ok {
		t.Errorf("%s's program was built with -tags=%s", archive, tags)
	}
	for _, kv := range [][2]string{{"CGO_ENABLED", "0"}, {"-trimpath", "true"}, {"vcs.modified", "false"}} {
		if got, ok := settings[kv[0]]; !ok || got != kv[1] {
			t.Errorf("%s's program: %s is %q, want %q", archive, kv[0], got, kv[1])
		}
	}
	rev := settings["vcs.revision"]
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(rev) {
		t.Errorf("%s's program: vcs.revision is %q, want a commit", archive, rev)
	} else if want := os.Getenv(distCommit); want != "" && rev != want {
		t.Errorf("%s's program: vcs.revision is %s, %s is %s", archive, rev, distCommit, want)
	}
}

// installFrom runs the unpacked archive's installer for this system to a scratch target with a scratch Bonsai home
// and checks the result: the target is the archive's program, install.json is right and check's reader reads it,
// and the installed --version names the version.
func installFrom(t *testing.T, release, version string) {
	tmp := t.TempDir()
	gitDir := ""
	if g, err := exec.LookPath("git"); err == nil {
		gitDir = filepath.Dir(g) // found before the PATH is narrowed on Windows
	}
	home := testpack.Isolate(t, tmp)
	root := testpack.Project(t, tmp, "project")
	if err := os.WriteFile(filepath.Join(root, "bonsai.yaml"),
		[]byte("format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(tmp, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", filepath.Join(tmp, "person"))

	var cmd *exec.Cmd
	var target, program string
	if runtime.GOOS == "windows" {
		ps := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if _, err := os.Stat(ps); err != nil {
			t.Fatalf("no Windows PowerShell at %s: %v", ps, err)
		}
		// Windows' own folders and git only on the PATH: no program of the person's (claude.exe among them).
		sys := os.Getenv("SystemRoot")
		t.Setenv("PATH", strings.Join([]string{filepath.Join(sys, "System32"), sys, filepath.Dir(ps), gitDir}, ";"))
		program, target = filepath.Join(release, "bonsai.exe"), filepath.Join(bin, "bonsai.exe")
		// A scratch target needs no rights, so install.ps1 raises no prompt there.
		cmd = exec.Command(ps, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File",
			filepath.Join(release, "install.ps1"), "-Target", target)
	} else {
		// A sudo that records it was called and fails, first on the PATH: a scratch target the test owns needs none.
		stubs := filepath.Join(tmp, "stubs")
		if err := os.Mkdir(stubs, 0o755); err != nil {
			t.Fatal(err)
		}
		called := filepath.Join(tmp, "sudo-called")
		if err := os.WriteFile(filepath.Join(stubs, "sudo"), []byte("#!/bin/sh\necho \"$@\" >>'"+called+"'\nexit 1\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", stubs+string(os.PathListSeparator)+os.Getenv("PATH"))
		defer func() {
			if _, err := os.Stat(called); err == nil {
				t.Error("install.sh called sudo for a scratch target")
			}
		}()
		program, target = filepath.Join(release, "bonsai"), filepath.Join(bin, "bonsai")
		cmd = exec.Command("sh", filepath.Join(release, "install.sh"), "--target", target)
	}
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("the installer from the archive: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "bonsai "+version) {
		t.Errorf("the installer's output does not name bonsai %s:\n%s", version, out)
	}
	if fileSum(t, target) != fileSum(t, program) {
		t.Fatal("the installed file is not the archive's program")
	}

	// install.json: the target's path as it resolves, the version, the program's SHA-256.
	var rec map[string]string
	raw, err := os.ReadFile(filepath.Join(home, engine.InstallFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("%s: %v", engine.InstallFile, err)
	}
	real, _ := engine.RealPath(target)
	if len(rec) != 3 || rec["path"] != filepath.ToSlash(real) || rec["version"] != version || rec["sha256"] != fileSum(t, program) {
		t.Errorf("%s is %v; want path %s, version %s and the program's sha256", engine.InstallFile, rec, filepath.ToSlash(real), version)
	}

	// check's reader, with the target first on the PATH: it reads the record and finds the bonsai right.
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	r, err := engine.Check(root, home)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range r.Findings {
		if f.Code == "bonsai-path" {
			t.Errorf("check found the installed bonsai wrong: %s", f.Sentence())
		}
	}
	for _, n := range r.Notes {
		if strings.Contains(n, engine.InstallFile) {
			t.Errorf("check did not read the record: %s", n)
		}
	}

	v, err := exec.Command(target, "--version").CombinedOutput()
	if err != nil || !strings.HasPrefix(string(v), "bonsai "+version+" ") {
		t.Errorf("the installed --version: %v %q; want bonsai %s", err, v, version)
	}
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
