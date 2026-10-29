package engine

// A newer release of Bonsai itself, known and handed over (plan-5 5.6.5; spec §3's 9 Oct note): the program stays a
// person's install, once per release on each side, and from 5.6 status and check say when a newer release exists, with
// the exact lines a person types to install it. Bonsai never installs or updates itself.
//
//   - What is read: the tags of Bonsai's public repository (ReleaseSource) with the machine's git (RemoteTags, the one
//     reader of tags): no login, no token, no service of Bonsai's. Only plain vX.Y.Z tags at or above 1.0.0 count
//     (NewestRelease): the old product's v0.x tags, pack tags such as base-v1.0.0 and pre-releases do not. A stamped
//     test build may name a local repository instead (the releaseSource stamp, honoured only in a stamped build).
//   - When: only bonsai status --full reads (RefreshRelease), at most once a day after a read that worked and once an
//     hour after one that failed, each read given ReleaseTimeout. The answer goes into the home's cache/release.json
//     (ReleaseFile, below), a refillable copy: deleting it costs one read. A write that fails never fails status: the
//     caller says in one line that it was not kept. Plain status, check and status --line read only that file.
//   - What is compared: the installed copy's version from the home's install.json, else the running binary's
//     (ComparedVersion). A version that is not three whole numbers (dev, or a pre-release) is not-a-release.
//   - What is said: status --json's bonsai_release (ReleaseRead.State), status's text line, check's warning
//     bonsai-newer, check's note when the record is missing or older than ReleaseStale, and --line's bonsai:vX.Y.Z.
//     No exit code changes for any of it.
//   - The lines (ReleaseLines): written from one table for both sides, built only from a version that parses as three
//     whole numbers, checked when the tags are read and again whenever the record is read, so a forged record gives
//     no lines. Every URL is ReleaseDownload, the checked version, then a fixed file name; the names follow
//     .goreleaser.yaml's archive and checksum templates, which a test reads as text.
//
// Its limits: a tag counts as a release whether or not its release is published (a tag waiting for approval, or one
// whose run failed before publishing, gives lines that fail to download); checking the release's own files would take
// an HTTPS request from Go, which would link the resolver the binary keeps out (TestTheBinaryLinksNoResolver), so
// that gap is left to 5.7's release path. The lines' fingerprint check proves the archive matches the release's own
// checksums.txt, not that the release is genuine; the line after it (gh attestation verify, 5.7.3) checks that, before
// anything from the archive runs.
//
// cache/release.json, Bonsai's own file, read by Bonsai alone (formats/README.md: not given a schema). Its fields:
//
//	newest  the newest release the last read that worked found, "1.0.1", or null (none read, or none at or above 1.0.0)
//	read    when that read worked, RFC 3339 in UTC, "2026-11-02T08:15:00Z", or null
//	source  the repository read, "https://github.com/LastStep/Bonsai.git" (a record of another source is read again)
//	failed  the last read, when it failed after the last that worked: {"at": "2026-11-03T08:15:00Z", "why": "git
//	        ls-remote failed: ..."}, else null

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Where Bonsai's releases are read from and downloaded, and who builds them: ReleaseRepo is the repository and
// ReleaseWorkflow the workflow the install lines' gh attestation verify holds an archive to.
const (
	ReleaseSource   = "https://github.com/LastStep/Bonsai.git"
	ReleaseDownload = "https://github.com/LastStep/Bonsai/releases/download/v"
	ReleaseRepo     = "LastStep/Bonsai"
	ReleaseWorkflow = "LastStep/Bonsai/.github/workflows/release.yml"
)

// ReleaseFile is the home's record of the last release read, relative to the home.
const ReleaseFile = "cache/release.json"

// How long a read may take, and how often one is made.
const (
	ReleaseTimeout = 10 * time.Second
	releaseEvery   = 24 * time.Hour // after a read that worked
	releaseRetry   = time.Hour      // after one that failed
	// ReleaseStale is how old a record may be before check's note asks for a read.
	ReleaseStale = 7 * 24 * time.Hour
)

// releaseSource is a test build's stamp (hookline.go): a local repository its release read names in place of
// ReleaseSource. Set only by the linker's -X, and honoured only in a stamped build; empty in every plain build.
var releaseSource string

// releaseNow is the clock the record's times come from; tests fake it.
var releaseNow = func() time.Time { return time.Now().UTC() }

// ReleaseFrom is the repository this build reads releases from.
func ReleaseFrom() string {
	if releaseSource != "" && Stamped() {
		return releaseSource
	}
	return ReleaseSource
}

// The release states, bonsai.status/1's bonsai_release.state (a closed list, its one home the schema).
const (
	ReleaseCurrent    = "current"
	ReleaseNewer      = "newer"
	ReleaseUnknown    = "unknown"
	ReleaseNotRelease = "not-a-release"
)

var plainVersion = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^([0-9]{1,9})\.([0-9]{1,9})\.([0-9]{1,9})$`) })

// PlainVersion parses a version of three whole numbers, X.Y.Z (digits and two dots, nothing else).
func PlainVersion(s string) ([3]int, bool) {
	var v [3]int
	m := plainVersion().FindStringSubmatch(s)
	if m == nil {
		return v, false
	}
	for i := range v {
		v[i], _ = strconv.Atoi(m[1+i])
	}
	return v, true
}

// NewestRelease is the newest release among tags, "" when none: only plain vX.Y.Z tags at or above 1.0.0 count.
func NewestRelease(tags []string) string {
	newest, have := [3]int{}, false
	for _, t := range tags {
		rest, ok := strings.CutPrefix(t, "v")
		if !ok {
			continue
		}
		v, ok := PlainVersion(rest)
		if !ok || v[0] < 1 {
			continue
		}
		if !have || older(newest, v) {
			newest, have = v, true
		}
	}
	if !have {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d", newest[0], newest[1], newest[2])
}

// ReleaseRecord is cache/release.json (this file's comment).
type ReleaseRecord struct {
	Newest *string         `json:"newest"`
	Read   *string         `json:"read"`
	Source string          `json:"source"`
	Failed *ReleaseFailure `json:"failed"`
}

// ReleaseFailure is the last read, when it failed.
type ReleaseFailure struct {
	At  string `json:"at"`
	Why string `json:"why"`
}

// ReleaseRead is the record as read: Record nil and Broken nil when there is none; Broken when it is not one Bonsai
// reads (Record then nil). Read says a read was made this run; NotKept why its answer could not be written.
type ReleaseRead struct {
	Record  *ReleaseRecord
	Broken  error
	Read    bool
	NotKept error
}

// ReleasePath is the record's full path in the home, forward slashes.
func ReleasePath(home string) string {
	return filepath.ToSlash(filepath.Join(home, filepath.FromSlash(ReleaseFile)))
}

// LoadRelease reads the home's record, checking every field: a newest that is not three whole numbers, a time that
// is not RFC 3339, an unknown shape, each makes the record broken, so nothing in it reaches a line.
func LoadRelease(home string) ReleaseRead {
	raw, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(ReleaseFile)))
	if errors.Is(err, fs.ErrNotExist) {
		return ReleaseRead{}
	}
	if err != nil {
		return ReleaseRead{Broken: err}
	}
	var rec ReleaseRecord
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rec); err != nil {
		return ReleaseRead{Broken: errors.New("it is not a JSON object of its four fields")}
	}
	if rec.Newest != nil {
		if _, ok := PlainVersion(*rec.Newest); !ok {
			return ReleaseRead{Broken: errors.New("its newest is not a version of three whole numbers")}
		}
	}
	for _, t := range []*string{rec.Read, failedAt(rec.Failed)} {
		if t != nil {
			if _, err := time.Parse(time.RFC3339, *t); err != nil {
				return ReleaseRead{Broken: errors.New("a time in it is not RFC 3339")}
			}
		}
	}
	if rec.Failed == nil && rec.Read == nil {
		return ReleaseRead{Broken: errors.New("it holds neither a read nor a failure")}
	}
	return ReleaseRead{Record: &rec}
}

func failedAt(f *ReleaseFailure) *string {
	if f == nil {
		return nil
	}
	return &f.At
}

// due says whether a read is due: no record, a broken one, one of another source, an hour after a failure, a day after
// a read that worked.
func (rr ReleaseRead) due(now time.Time) bool {
	rec := rr.Record
	if rec == nil || rec.Source != ReleaseFrom() {
		return true
	}
	if rec.Failed != nil {
		at, _ := time.Parse(time.RFC3339, rec.Failed.At)
		return !now.Before(at.Add(releaseRetry))
	}
	read, _ := time.Parse(time.RFC3339, *rec.Read)
	return !now.Before(read.Add(releaseEvery))
}

// RefreshRelease is status --full's read: when one is due, the source's tags (tags, given ReleaseTimeout), the
// answer written to the home's record. A read that fails keeps the last one that worked and records the failure. It
// gives the record as it now stands; NotKept says why it could not be written, which never fails the caller.
func RefreshRelease(home string, tags func(source string, timeout time.Duration) ([]string, error)) ReleaseRead {
	rr := LoadRelease(home)
	now := releaseNow()
	if !rr.due(now) {
		return rr
	}
	rec := ReleaseRecord{Source: ReleaseFrom()}
	if rr.Record != nil && rr.Record.Source == rec.Source {
		rec.Newest, rec.Read = rr.Record.Newest, rr.Record.Read
	}
	stamp := now.UTC().Format(time.RFC3339)
	if list, err := tags(rec.Source, ReleaseTimeout); err != nil {
		rec.Failed = &ReleaseFailure{At: stamp, Why: ascii(oneLine(err.Error()))}
	} else {
		rec.Newest, rec.Read = nil, &stamp
		if n := NewestRelease(list); n != "" {
			rec.Newest = &n
		}
	}
	out := ReleaseRead{Record: &rec, Read: true}
	raw, err := json.MarshalIndent(rec, "", "  ")
	if err == nil {
		err = workspace.WriteFileAtomic(filepath.Join(home, filepath.FromSlash(ReleaseFile)), append(raw, '\n'))
	}
	out.NotKept = err
	return out
}

// ComparedVersion is the version a release is compared with: the installed copy's, from the home's install.json,
// else the running binary's.
func ComparedVersion(home, running string) string {
	if in, err := readInstall(home); err == nil && in != nil && in.Version != "" {
		return in.Version
	}
	return running
}

// State is status --json's bonsai_release for the version compared (ComparedVersion's), from the record alone.
func (rr ReleaseRead) State(home, version string) *format.StatusRelease {
	s := &format.StatusRelease{State: ReleaseUnknown}
	if version != "" && version != "dev" {
		v := ascii(version)
		s.Version = &v
	}
	why := func(w string) *format.StatusRelease { s.Why = &w; return s }
	switch {
	case rr.Broken != nil:
		return why("the Bonsai home's " + ReleasePath(home) + " is not a record Bonsai reads (" + ascii(oneLine(rr.Broken.Error())) +
			"), so nothing in it is used: bonsai status --full reads the releases again")
	case rr.Record == nil:
		return why("Bonsai's releases have not been read on this machine yet: bonsai status --full reads them")
	}
	rec := rr.Record
	s.Newest, s.Read = rec.Newest, rec.Read
	if rec.Failed != nil {
		return why("the last read of Bonsai's releases, at " + rec.Failed.At + ", failed: " + ascii(oneLine(rec.Failed.Why)))
	}
	have, ok := PlainVersion(version)
	switch {
	case !ok:
		s.State = ReleaseNotRelease
	case rec.Newest == nil:
		s.State = ReleaseCurrent
	default:
		newest, _ := PlainVersion(*rec.Newest)
		s.State = ReleaseCurrent
		switch {
		case older(have, newest):
			s.State, s.Lines = ReleaseNewer, ReleaseLines(*rec.Newest, runtime.GOARCH)
		case older(newest, have):
			// BUG-139: the record names an older release than this bonsai (it was read before this one was installed),
			// so it says nothing of what is newest now: unknown, never current.
			s.State = ReleaseUnknown
			return why("the home's record names v" + *rec.Newest + ", an older release than this bonsai's v" + ascii(version) +
				", as read at " + *rec.Read + ", so it cannot say whether a newer one exists: bonsai status --full reads the releases again")
		}
	}
	return s
}

// releaseLines is the one table of the install lines, for a release V on architecture A (each side's archive named
// as .goreleaser.yaml's archives name_template makes it, bonsai_V_<os>_A, and checksum's checksums.txt): bash for
// Linux and WSL, PowerShell 5.1 for Windows, one command per line, no &&, run one at a time: seven lines a side,
// every part fixed but the checked version. The fourth line prints OK (PowerShell: True) when the archive matches the
// release's checksums.txt. The fifth, the same words on both sides, exits 0 and prints Verification succeeded only when
// the archive was built by ReleaseWorkflow in ReleaseRepo, for the tag v<version>, on GitHub's own machines
// (--deny-self-hosted-runners: without it a record made on anyone's own machine would pass); it needs gh 2.97 or
// newer, logged in (no login exits 4, the wrong origin 1). It comes before unpacking because the installer is inside
// the archive and runs as root or administrator, so nothing from the archive runs before its origin is checked.
// Anything else from either, the person stops. The installer sits at the archive's top beside the binary (5.7 puts it
// there). docs/install.md shows these lines for 1.0.1, held to this table by internal/selfcheck's test.
var releaseLines = struct{ linux, windows []string }{
	linux: []string{
		`cd "$(mktemp -d)"`,
		`curl -fsSLO {url}/{archive}`,
		`curl -fsSLO {url}/checksums.txt`,
		`sha256sum -c --ignore-missing checksums.txt`,
		releaseVerify,
		`tar -xzf {archive}`,
		`sh install.sh`,
	},
	windows: []string{
		`cd (New-Item -ItemType Directory (Join-Path $env:TEMP ("bonsai-" + [guid]::NewGuid())))`,
		`curl.exe -fsSLO {url}/{archive}`,
		`curl.exe -fsSLO {url}/checksums.txt`,
		`(Get-FileHash {archive} -Algorithm SHA256).Hash -eq ((Select-String -Path checksums.txt -SimpleMatch {archive}).Line -split ' ')[0]`,
		releaseVerify,
		`Expand-Archive {archive} -DestinationPath . -Force`,
		`powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1`,
	},
}

// releaseVerify is the fifth line on each side, the same words in bash and PowerShell.
const releaseVerify = `gh attestation verify {archive} -R ` + ReleaseRepo + ` --signer-workflow ` + ReleaseWorkflow +
	` --source-ref refs/tags/v{version} --deny-self-hosted-runners`

// ReleaseArchive is a release's archive name on one side, as .goreleaser.yaml names it.
func ReleaseArchive(version, goos, arch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return "bonsai_" + version + "_" + goos + "_" + arch + ext
}

// ReleaseLines are the lines that install release version on each side, for this machine's architecture; nil unless
// version is three whole numbers, the only part of a line not fixed here.
func ReleaseLines(version, arch string) *format.ReleaseLines {
	if _, ok := PlainVersion(version); !ok || (arch != "amd64" && arch != "arm64") {
		return nil
	}
	fill := func(lines []string, goos string) []string {
		out := make([]string, len(lines))
		r := strings.NewReplacer("{url}", ReleaseDownload+version, "{archive}", ReleaseArchive(version, goos, arch), "{version}", version)
		for i, l := range lines {
			out[i] = r.Replace(l)
		}
		return out
	}
	return &format.ReleaseLines{Linux: fill(releaseLines.linux, "linux"), Windows: fill(releaseLines.windows, "windows")}
}

// checkRelease gives check's word on Bonsai's releases, from the home's record alone: the warning bonsai-newer when a
// newer release is recorded; a note naming status --full when the record is missing, broken or older than
// ReleaseStale. Nothing when the version is not a release's and the record is fresh.
func (r *CheckResult) checkRelease() {
	rr := LoadRelease(r.Home)
	s := rr.State(r.Home, ComparedVersion(r.Home, r.version))
	if s.State == ReleaseNewer {
		r.add("bonsai-newer", "", "", "Bonsai v"+*s.Newest+" is released, and this machine's Bonsai is v"+*s.Version+" (as bonsai status --full read "+
			"the releases at "+*s.Read+")", "a person installs Bonsai v"+*s.Newest+" on each side: the lines are in `bonsai status --full --json`, "+
			"`bonsai_release.lines`")
	}
	switch {
	case rr.Broken != nil || rr.Record == nil:
		r.Notes = append(r.Notes, ascii(*s.Why))
	case rr.Record.Read == nil:
		r.Notes = append(r.Notes, "Bonsai's releases have not been read on this machine yet (the last read failed): bonsai status --full reads them")
	default:
		read, _ := time.Parse(time.RFC3339, *rr.Record.Read)
		if releaseNow().Sub(read) > ReleaseStale {
			r.Notes = append(r.Notes, "Bonsai's releases were last read at "+*rr.Record.Read+", over 7 days ago: bonsai status --full reads them again")
		}
	}
}
