package agents

// The agents folder as the cleaner reads it (internal/clean's agents kind; format.GeneratedKinds' agents): its session
// folders, the files Bonsai writes in each, the time of a folder's newest line, and the brief ties. Only Bonsai's own
// names are listed, each a regular file in a plain folder (a link is never followed), so nothing else there is ever
// named for deleting.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Plain reports whether a path is a plain folder: not a link, a junction or a file.
func Plain(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode().Type() == fs.ModeDir
}

// Sessions lists the session folders, by name: each a plain folder whose name is a session id (the ties' folder
// apart).
func (s Store) Sessions() ([]string, error) {
	entries, err := os.ReadDir(s.Dir())
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.Name() != TiesFolder && e.Type() == fs.ModeDir && validName(e.Name()) == nil {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// SessionFiles lists the files Bonsai writes in a session's folder, by their paths inside it (forward slashes), each a
// regular file: claims/<call> first, then spawns.ndjson, then each agent's <agent>.calls and <agent>.state. Anything
// else in the folder is not named.
func (s Store) SessionFiles(session string) ([]string, error) {
	if err := validName(session); err != nil {
		return nil, err
	}
	dir := s.SessionDir(session)
	var out []string
	regular := func(p string) bool {
		fi, err := os.Lstat(p)
		return err == nil && fi.Mode().IsRegular()
	}
	if claims := filepath.Join(dir, ClaimsFolder); Plain(claims) {
		entries, err := os.ReadDir(claims)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if validName(e.Name()) == nil && regular(filepath.Join(claims, e.Name())) {
				out = append(out, ClaimsFolder+"/"+e.Name())
			}
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var agentFiles []string
	for _, e := range entries {
		name := e.Name()
		if !regular(filepath.Join(dir, name)) {
			continue
		}
		switch {
		case name == spawnsFile:
			out = append(out, name)
		case strings.HasSuffix(name, callsExt) && validName(strings.TrimSuffix(name, callsExt)) == nil,
			strings.HasSuffix(name, stateExt) && validName(strings.TrimSuffix(name, stateExt)) == nil:
			agentFiles = append(agentFiles, name)
		}
	}
	sort.Strings(agentFiles) // <agent>.calls before <agent>.state
	return append(out, agentFiles...), nil
}

// LastLine gives the time of a session folder's newest line: each agent's last call line, each state's last event
// (or its header's first call), the last spawn call, and a claim's modification time (a claim holds no time); a file
// whose own time does not read counts by its modification time. ok is false when the folder holds none of Bonsai's
// files.
func (s Store) LastLine(session string) (newest time.Time, ok bool) {
	files, err := s.SessionFiles(session)
	if err != nil {
		return time.Time{}, false
	}
	dir := s.SessionDir(session)
	note := func(t time.Time) {
		if !ok || t.After(newest) {
			newest, ok = t, true
		}
	}
	mtime := func(rel string) {
		if fi, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			note(fi.ModTime())
		}
	}
	for _, f := range files {
		switch {
		case strings.HasSuffix(f, callsExt):
			if at, _, read := s.LastCall(session, strings.TrimSuffix(f, callsExt)); read {
				note(at)
				continue
			}
		case strings.HasSuffix(f, stateExt):
			if st, err := s.Read(session, strings.TrimSuffix(f, stateExt)); err == nil {
				at := st.Header.First
				if len(st.Events) > 0 {
					at = st.Events[len(st.Events)-1].At
				}
				if t, err := ParseStamp(at); err == nil {
					note(t)
					continue
				}
			}
		case f == spawnsFile:
			if sp, err := s.Spawns(session); err == nil && len(sp) > 0 {
				if t, err := ParseStamp(sp[len(sp)-1].At); err == nil {
					note(t)
					continue
				}
			}
		}
		mtime(f)
	}
	return newest, ok
}

// Tie is one brief's tie file: the brief and the agent it names (session and agent; "" when it names none).
type Tie struct {
	Brief, Session, Agent string
}

// Ties lists the tie files, by brief: each a regular file named a brief's id.
func (s Store) Ties() ([]Tie, error) {
	if !Plain(s.Dir()) || !Plain(filepath.Join(s.Dir(), TiesFolder)) {
		return nil, nil
	}
	entries, err := os.ReadDir(filepath.Join(s.Dir(), TiesFolder))
	if err != nil {
		return nil, err
	}
	var out []Tie
	for _, e := range entries {
		if validName(e.Name()) != nil || !e.Type().IsRegular() {
			continue
		}
		t := Tie{Brief: e.Name()}
		t.Session, t.Agent, _ = s.TiedTo(e.Name())
		out = append(out, t)
	}
	return out, nil
}

// TiePath is a tie file's path.
func (s Store) TiePath(brief string) string { return filepath.Join(s.Dir(), TiesFolder, brief) }
