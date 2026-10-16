package engine

// check's findings and warnings on this machine (contract §3; spec §3): its record of the checkout in the Bonsai home
// (workspace.json, written by init and update: workspace/record.go) and the bonsai on the PATH.
//
//   - id-changed: bonsai.yaml's id is not the last one this machine recorded for the main checkout (contract §3: "a
//     changed id is a problem in bonsai check and status --json"). A checkout this machine has no record of (a fresh
//     clone, CI) has nothing to compare, and gives nothing;
//   - same-id (a warning): another checkout on this machine holds the same id: another machine folder's record whose
//     last id is this one, at a path that still holds a bonsai.yaml with it (a record whose checkout moved or changed
//     its id is left alone: that is the stranded folder's, below);
//   - stranded (a warning; step 5.6.1, contract section 3 "Moved checkouts"): another machine folder whose record's
//     last id is this one, whose recorded path holds no bonsai.yaml at all (the checkout moved or is gone), and which
//     still holds settings or attached labels, which this checkout no longer gets. Its next step is a person's exact
//     lines that bring them over and clear the old folder (Stranded);
//   - machine-record: a record of a checkout in the Bonsai home that Bonsai does not read (workspace/record.go): one
//     such record un-verifies every main checkout on the machine (FirstClaim cannot name the first claim while it might
//     hold it), so the guard honours no task's grant anywhere here until a person mends it (step 5.3.1's verifier);
//   - bonsai-path: the bonsai found on the PATH is not the installed one (spec §3, "A tripwire, not a wall"): the home's
//     install.json, which Bonsai's installer writes from step 5.6 (its path, version and SHA-256, spec §3), names
//     another file, or the installed file is not the one it recorded. With no install.json there is nothing to compare
//     with, and check says so in a note to a person, never a finding or a warning.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// InstallFile is the home's record of the installed bonsai (spec §3, §10), written by Bonsai's installer (step 5.6).
const InstallFile = "install.json"

// checkMachine checks this machine's record of the checkout and the bonsai on the PATH.
func (r *CheckResult) checkMachine() {
	if r.Home == "" {
		return
	}
	r.checkMachineRecords()
	id := r.Config.ID
	if rec, err := workspace.LoadMachineRecord(r.Home, r.Main); err == nil && rec != nil && rec.Current() != "" && rec.Current() != id && id != "" {
		r.add("id-changed", workspace.ConfigFile, "", "bonsai.yaml's id changed from "+rec.Current()+" to "+ascii(id)+
			" (this machine's record of the checkout; the studio stops forwarding the project until it is registered again)",
			"if the change was not meant, a person puts bonsai.yaml's id back to "+rec.Current()+"; if it was (a copy made its "+
				"own project), record the new id on this machine, run: bonsai update --yes")
	}
	if id != "" {
		key, _ := workspace.MachineKey(r.Main)
		for _, kr := range workspace.MachineRecords(r.Home) {
			if kr.Key == key || kr.Record.Current() != id {
				continue
			}
			other := filepath.FromSlash(kr.Record.Path)
			if samePath(other, r.Main) {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(other, workspace.ConfigFile))
			if err != nil {
				continue
			}
			cfg, err := workspace.ReadConfig(raw)
			if err != nil || cfg.ID != id {
				continue
			}
			r.add("same-id", workspace.ConfigFile, "", "another checkout on this machine, "+ascii(kr.Record.Path)+", holds this workspace's id "+id,
				"a clone of the same project is fine (every clone shares the id): leave both; a copy meant as a new project needs its "+
					"own id: in that copy's main checkout, run: bonsai init --new-id")
		}
	}
	if id != "" {
		for _, sf := range Stranded(r.Home, r.Main, id, runtime.GOOS) {
			r.add("stranded", "", "", sf.Message, sf.Next)
		}
	}
	r.checkInstalled()
}

// StrandedFolder is one machine folder stranded under an old path: what it holds, and the person's lines.
type StrandedFolder struct {
	Folder     string   // <home>/workspaces/r-<16 hex>, forward slashes
	OldPath    string   // the checkout's path its record holds, forward slashes
	Command    string   // its status_command when its status_writes is command and the file keeps its rules; "" else
	Unread     string   // why its settings.json is not read, "" when it is (or there is none)
	Namespaces []string // the namespaces attached there
	Lines      []string // the person's lines, in order: bring each over, then remove the folder
	Message    string   // check's sentence
	Next       string   // check's next step (who: person)
}

// machineFolderName is a machine folder's name as Bonsai makes it: the only name a removal line is built from.
var machineFolderName = regexp.MustCompile(`^r-[0-9a-f]{16}$`)

// Stranded lists the home's machine folders stranded under an old path for the workspace id whose main checkout
// is main (contract section 3, "Moved checkouts"), by folder name: each other folder whose record holds id as its
// last, whose recorded path holds no bonsai.yaml (a path holding one with another id is that checkout's own folder,
// whose check gives id-changed), and which holds a settings.json or attached labels (a folder holding only its record
// loses nothing). goos says which shell the lines are for: PowerShell on windows, a POSIX shell else.
func Stranded(home, main, id, goos string) []StrandedFolder {
	if home == "" || id == "" {
		return nil
	}
	own, _ := workspace.MachineKey(main)
	var out []StrandedFolder
	for _, kr := range workspace.MachineRecords(home) {
		if kr.Key == own || kr.Record.Current() != id {
			continue
		}
		old := filepath.FromSlash(kr.Record.Path)
		if samePath(old, main) {
			continue
		}
		if _, err := os.Lstat(filepath.Join(old, workspace.ConfigFile)); !errors.Is(err, fs.ErrNotExist) {
			continue // a bonsai.yaml there (or a path that cannot be looked at): not stranded
		}
		dir := filepath.Join(home, "workspaces", kr.Key)
		sf := StrandedFolder{Folder: filepath.ToSlash(dir), OldPath: kr.Record.Path, Namespaces: workspace.AttachedIn(dir)}
		_, settingsErr := os.Lstat(filepath.Join(dir, workspace.MachineSettingsFile))
		hasSettings := settingsErr == nil
		if !hasSettings && len(sf.Namespaces) == 0 {
			continue
		}
		if hasSettings {
			ms, err := workspace.MachineSettingsIn(dir)
			switch {
			case err != nil:
				sf.Unread = oneLine(err.Error())
			case ms.StatusWrites == workspace.StatusWritesCommand:
				sf.Command = ms.StatusCommand
			}
		}
		sf.describe(home, main, kr.Key, goos)
		out = append(out, sf)
	}
	return out
}

// describe writes the folder's sentence, against what this checkout has now, and the person's lines.
func (sf *StrandedFolder) describe(home, main, key, goos string) {
	quote := shQuote
	if goos == "windows" {
		quote = psQuote
	}
	var held []string
	switch {
	case sf.Unread != "":
		held = append(held, "a settings.json Bonsai does not read ("+sf.Unread+")")
	case sf.Command != "":
		held = append(held, "status_writes command, its status_command "+sf.Command)
	}
	if len(sf.Namespaces) > 0 {
		held = append(held, "the attached labels "+strings.Join(sf.Namespaces, ", "))
	}
	if len(held) == 0 {
		held = append(held, "a settings.json with status_writes agents, the default")
	}
	now := "status_writes agents"
	if ms, err := workspace.LoadMachineSettings(home, main); err != nil {
		now = "settings Bonsai does not read"
	} else if ms.StatusWrites == workspace.StatusWritesCommand {
		now = "status_writes command (" + ms.StatusCommand + ")"
	}
	if dir, err := workspace.MachineDir(home, main); err == nil {
		if ns := workspace.AttachedIn(dir); len(ns) > 0 {
			now += " and the attached labels " + strings.Join(ns, ", ")
		} else {
			now += " and no attached labels"
		}
	}
	sf.Message = "a machine folder of this workspace, " + sf.Folder + ", is stranded under its old checkout " + sf.OldPath +
		", which holds no bonsai.yaml now (moved or gone): it still holds " + strings.Join(held, "; ") +
		", which this checkout does not get; this checkout has " + now
	var steps []string
	if sf.Unread != "" {
		steps = append(steps, "read "+sf.Folder+"/"+workspace.MachineSettingsFile+" and set what it means to hold here with bonsai settings set")
	} else if sf.Command != "" {
		sf.Lines = append(sf.Lines, `bonsai settings set status_writes=command "status_command=`+sf.Command+`"`)
	}
	for _, ns := range sf.Namespaces {
		sf.Lines = append(sf.Lines, "bonsai labels attach "+quote(sf.Folder+"/"+workspace.MachineLabelsDir+"/"+ns+".yaml"))
	}
	// The removal's path is built from the folder's own name alone, once it is a machine folder's, directly under
	// <home>/workspaces/: never from the record, never the home or workspaces/ itself.
	if machineFolderName.MatchString(key) {
		gone := filepath.ToSlash(filepath.Join(home, "workspaces", key))
		if goos == "windows" {
			sf.Lines = append(sf.Lines, "Remove-Item -Recurse -LiteralPath "+quote(gone))
		} else {
			sf.Lines = append(sf.Lines, "rm -r "+quote(gone))
		}
	}
	shell := "a shell"
	if goos == "windows" {
		shell = "PowerShell"
	}
	if len(sf.Lines) > 0 {
		steps = append(steps, "run in this checkout, in "+shell+", in order: "+strings.Join(sf.Lines, " ; "))
	}
	sf.Next = "a person, in a terminal of their own: " + strings.Join(steps, "; then ") +
		"; or, in a workspace a studio manages, registers this checkout with the studio again (contract section 15.2)"
}

// shQuote quotes a word for a POSIX shell: single quotes, each quote inside it ends the quoting, is escaped with a
// backslash, and starts it again.
func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// psQuote quotes a word for PowerShell: single quotes, a ' inside doubled.
func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// checkMachineRecords gives a machine-record finding for each record of a checkout in the home that Bonsai does not
// read (a folder workspaces/r-* holding a workspace.json that workspace.MachineRecords leaves out), in the folders'
// order, naming the file, why when FirstClaim names it, and the person's step.
func (r *CheckResult) checkMachineRecords() {
	dir := filepath.Join(r.Home, "workspaces")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	read := map[string]bool{}
	for _, kr := range workspace.MachineRecords(r.Home) {
		read[kr.Key] = true
	}
	var named *workspace.Error
	if _, _, err := workspace.FirstClaim(r.Home, ""); err != nil {
		_ = errors.As(err, &named)
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "r-") || read[e.Name()] {
			continue
		}
		p := filepath.Join(dir, e.Name(), workspace.MachineRecordFile)
		if _, err := os.Lstat(p); err != nil {
			continue // a folder with no record holds nothing to verify
		}
		show := filepath.ToSlash(p)
		why := "is not a record Bonsai reads"
		next := "a person checks " + show + " (this machine's record of a checkout) and deletes it; bonsai update in that checkout " +
			"then writes it again, its claim on the id starting then"
		if named != nil && named.File == show {
			why, next = named.Msg, named.Next
		}
		r.add("machine-record", "", "", "this machine's record "+ascii(show)+" in Bonsai's home "+ascii(why)+": until it is mended, no main "+
			"checkout on this machine is verified, so the guard honours no task's grant in any project here (main-not-verified)",
			ascii(next))
	}
}

// install is the home's install.json as spec §3 names its fields.
type install struct {
	Path, Version, SHA256 string
}

// readInstall reads <home>/install.json: nil when there is none.
func readInstall(home string) (*install, error) {
	p := filepath.Join(home, InstallFile)
	raw, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	o, ok := v.(schema.Object)
	if err != nil || !ok || o.String("path") == "" {
		return nil, errors.New("it is not a JSON object with the installed file's path")
	}
	return &install{Path: o.String("path"), Version: o.String("version"), SHA256: strings.ToLower(o.String("sha256"))}, nil
}

// checkInstalled compares the bonsai on the PATH with the installed one (spec §3).
func (r *CheckResult) checkInstalled() {
	in, err := readInstall(r.Home)
	switch {
	case err != nil:
		r.add("bonsai-path", "", "", "the Bonsai home's "+InstallFile+" (the installed bonsai's record) is not one Bonsai reads: "+
			ascii(oneLine(err.Error())), "a person installs Bonsai's release again, which writes the record (spec section 3: the "+
			"program is a person's install), then run: bonsai check")
		return
	case in == nil:
		r.Notes = append(r.Notes, "no "+InstallFile+" in the Bonsai home yet (Bonsai's installer writes it from step 5.6): the bonsai "+
			"on the PATH was not compared with an installed one")
		return
	}
	found, err := exec.LookPath("bonsai")
	if err != nil {
		r.Notes = append(r.Notes, "no bonsai on the PATH, so none was compared with the installed one, "+ascii(in.Path))
		return
	}
	if abs, err := filepath.Abs(found); err == nil {
		found = abs
	}
	installed := filepath.FromSlash(in.Path)
	if !samePath(found, installed) {
		r.add("bonsai-path", "", "", "the bonsai on the PATH, "+ascii(filepath.ToSlash(found))+", is not the installed one, "+ascii(in.Path)+
			" (install.json): a hook or command that calls bonsai by name runs it", "a person takes "+ascii(filepath.ToSlash(found))+
			" off the PATH (deletes it, or puts "+ascii(filepath.ToSlash(filepath.Dir(installed)))+" first on the PATH), then run: bonsai check")
		return
	}
	if in.SHA256 != "" {
		if sum, err := fileSHA256(found); err == nil && sum != in.SHA256 {
			r.add("bonsai-path", "", "", "the installed bonsai, "+ascii(in.Path)+", is not the file its install recorded (its SHA-256 differs "+
				"from install.json's)", "a person installs Bonsai's release "+ascii(in.Version)+" again (spec section 3: the program is a "+
				"person's install), then run: bonsai check")
		}
	}
}

func fileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// RecordCheckout writes this machine's record of the plan's main checkout (workspace/record.go) once init or update
// ended without a refusal, in the main checkout only: its path, bonsai.yaml's id, the branch it holds as the
// workspace's base, and the settings for the checks the plan's bonsai.yaml holds, the ones the command read and judged
// (another base, or looser settings, than the ones recorded are refused to an agent before this: BaseChange,
// ChecksLooser). It reports what went wrong, for the command's text, and never fails the command (the project's files
// are written).
func RecordCheckout(p *Plan) error {
	if p == nil || p.Config == nil || p.Home == "" || !samePath(p.Root, p.Main) {
		return nil
	}
	_, err := workspace.RecordCheckoutWith(p.Home, p.Main, p.Config.ID, recordClock(), workspace.ChecksOf(p.Config.Full))
	return err
}

// ChecksLooser says how the settings for the checks in the plan's bonsai.yaml, which init or update records once it
// ends (RecordCheckout), are looser than the ones this machine's record of the main checkout accepted
// (workspace.Loosenings): each loosening, in words. None when they are the same or stricter, when there is no record
// or it holds none yet (that first record is free, as the base's is), or when the plan's checkout is not the main one
// (only the main checkout records). Recording looser settings is a person's step (workspace/checks.go: an agent could
// commit a lowered floor on the base branch and record it, which would make the ladder's check moot), which the
// command refuses in an agent's session (cmd/bonsai).
func ChecksLooser(p *Plan) []string {
	if p == nil || p.Config == nil || p.Home == "" || !samePath(p.Root, p.Main) {
		return nil
	}
	rec, err := workspace.LoadMachineRecord(p.Home, p.Main)
	if err != nil || rec == nil || rec.Checks == nil {
		return nil
	}
	return workspace.Loosenings(rec.Checks, workspace.ChecksOf(p.Config.Full))
}

// BaseChange says whether this machine's record of the plan's main checkout, written once init or update ends
// (RecordCheckout), would record a base branch other than the one it already names: was and now, and changes true
// when they differ. Recording another base is a person's step (workspace/record.go: an agent could switch the main
// checkout to a branch whose bonsai.yaml says otherwise and record it), which the command refuses in an agent's
// session (cmd/bonsai). The first record, a record that names no base yet, the same branch again, and a detached HEAD
// (which records none) are no change. Only the main checkout records.
func BaseChange(p *Plan) (was, now string, changes bool) {
	if p == nil || p.Config == nil || p.Home == "" || !samePath(p.Root, p.Main) {
		return "", "", false
	}
	rec, err := workspace.LoadMachineRecord(p.Home, p.Main)
	if err != nil || rec == nil || rec.BaseBranch == "" {
		return "", "", false
	}
	now = workspace.ReadGitLayout(p.Main).Branch()
	if now == "" || now == rec.BaseBranch {
		return "", "", false
	}
	return rec.BaseBranch, now, true
}

// recordClock is the time a machine record is stamped with (a variable, so a test can set it).
var recordClock = func() time.Time { return time.Now() }
