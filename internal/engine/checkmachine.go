package engine

// check's findings and warnings on this machine (contract §3; spec §3): its record of the checkout in the Bonsai home
// (workspace.json, written by init and update: workspace/record.go) and the bonsai on the PATH.
//
//   - id-changed: bonsai.yaml's id is not the last one this machine recorded for the main checkout (contract §3: "a
//     changed id is a problem in bonsai check and status --json"). A checkout this machine has no record of (a fresh
//     clone, CI) has nothing to compare, and gives nothing;
//   - same-id (a warning): another checkout on this machine holds the same id: another machine folder's record whose
//     last id is this one, at a path that still holds a bonsai.yaml with it (a record whose checkout moved or changed
//     its id is left alone: the stranded folder is step 5.6's);
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
	id := r.Config.ID
	if rec, err := workspace.LoadMachineRecord(r.Home, r.Main); err == nil && rec != nil && rec.Current() != "" && rec.Current() != id && id != "" {
		r.add("id-changed", workspace.ConfigFile, "", "bonsai.yaml's id changed from "+rec.Current()+" to "+ascii(id)+
			" (this machine's record of the checkout; the studio stops forwarding the project until it is registered again)",
			"if the change was not meant, a person puts bonsai.yaml's id back to "+rec.Current()+"; if it was (a copy made its "+
				"own project), run: bonsai update --yes (it records the new id on this machine)")
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
	r.checkInstalled()
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
	shown := ascii(filepath.ToSlash(filepath.Join(r.Home, InstallFile)))
	in, err := readInstall(r.Home)
	switch {
	case err != nil:
		r.add("bonsai-path", "", "", shown+" (the installed bonsai's record) is not one Bonsai reads: "+ascii(oneLine(err.Error())),
			"a person installs Bonsai's release again, which writes the record (spec section 3: the program is a person's install)")
		return
	case in == nil:
		r.Notes = append(r.Notes, "no "+shown+" (Bonsai's installer writes it from step 5.6), so the bonsai on the PATH was not "+
			"compared with an installed one")
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
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// RecordCheckout writes this machine's record of the plan's main checkout (workspace/record.go) once init or update
// ended without a refusal, in the main checkout only: its path and bonsai.yaml's id. It reports what went wrong, for
// the command's text, and never fails the command (the project's files are written).
func RecordCheckout(p *Plan) error {
	if p == nil || p.Config == nil || p.Home == "" || !samePath(p.Root, p.Main) {
		return nil
	}
	_, err := workspace.RecordCheckout(p.Home, p.Main, p.Config.ID, recordClock())
	return err
}

// recordClock is the time a machine record is stamped with (a variable, so a test can set it).
var recordClock = func() time.Time { return time.Now() }
