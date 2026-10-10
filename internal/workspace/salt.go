package workspace

// The home's salt (contract §3: "the input-hash key; never leaves the machine"): <home>/salt, 64 lower-case hex
// characters from 32 random bytes, no line ending. The recorder keys a log record's input_hash with it (contract
// §8.1; step 5.2.4), so the records of one tool call pair up without the call's input being kept. It is this
// machine's alone: no format carries it, no command prints it, and base's walls deny reading it (spec §7; step 5.5).
//
// It is made at first need. The whole file is written under a temporary name beside it and then hard-linked into
// place: a link never replaces a file, so of two first writers exactly one wins, the other reads the winner's, and
// no reader ever sees half a file. Mode 0600 on Linux and macOS (Windows keeps the home's own permissions).

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SaltFile is the salt's name in the Bonsai home.
const SaltFile = "salt"

// link puts a file in place under a second name, failing when the name is taken (a variable for the tests).
var link = os.Link

// Salt gives this machine's salt, making it at first need (this file's comment). A caller that gets an error writes
// its record with input_hash null: a record is never lost for want of the salt (design/plan-5.md, 5.2.2 note 5).
func Salt() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return saltAt(filepath.Join(home, SaltFile))
}

// saltAt reads the salt at path, making it when it is missing.
func saltAt(path string) (string, error) {
	s, err := readSalt(path)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		return s, err
	}
	if err := makeSalt(path); err != nil {
		return "", &Error{File: filepath.ToSlash(path), Msg: "cannot be made: " + oneLine(err), Err: err,
			Next: "check that the Bonsai home can be written (bonsai status names it), then run the command again"}
	}
	return readSalt(path)
}

// makeSalt writes a new salt beside path and links it into place. Losing the race to another writer is no error:
// the winner's salt is there.
func makeSalt(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var key [32]byte
	if _, err := rand.Read(key[:]); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+SaltFile+".tmp-*") // mode 0600
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() { _ = RemoveFile(tmp) }()
	if _, err := f.WriteString(hex.EncodeToString(key[:])); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := link(tmp, path); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return nil
}

// readSalt reads the salt at path and holds it to its shape: 64 lower-case hex characters (a final line ending, which
// an editor may add, is let through). A missing file is an error whose errors.Is(err, fs.ErrNotExist) holds.
func readSalt(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		return "", &Error{File: filepath.ToSlash(path), Msg: "cannot be read: " + oneLine(err), Err: err,
			Next: "check the file's permissions in the Bonsai home, then run the command again"}
	}
	s := strings.TrimSuffix(strings.TrimSuffix(string(raw), "\n"), "\r")
	if len(s) != 64 || strings.Trim(s, "0123456789abcdef") != "" {
		return "", &Error{File: filepath.ToSlash(path), Msg: "is not 64 lower-case hex characters, so it is not used",
			Next: "delete the file; Bonsai makes a new salt at the next record (records written before it no longer " +
				"pair with records written after)"}
	}
	return s, nil
}
