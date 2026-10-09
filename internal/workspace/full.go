package workspace

// bonsai.yaml read in full (plan-5 5.1.4a): the engine and bonsai check read every field of bonsai.workspace/1,
// held to its schema (internal/format), beside ReadConfig's fields. The guard does not: it reads bonsai.yaml on
// every hook call through LoadConfig and ReadConfig (config.go), which read only the fields the walking skeleton's
// guard and engine use, so a field the guard does not judge by (a ladder rung, a generated kind) never blocks an
// edit, and a hook call never pays for the full read (spec §3: a hook starts in under 5 ms). This file adds to
// config.go and changes none of it; TestGuardReadIsLean holds the guard's read to its fields.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/format"
)

// ReadConfigFull reads bonsai.yaml's bytes in full: ReadConfig's read, with its checks and messages, then every
// field held to bonsai.workspace/1's schema and read into Config.Full. Every key is kept, as ReadConfig keeps it.
func ReadConfigFull(raw []byte) (*Config, error) {
	c, err := ReadConfig(raw)
	if err != nil {
		return nil, err
	}
	w, err := format.WorkspaceFromMap(c.Doc)
	if err != nil {
		return nil, fileError(ConfigFile, err)
	}
	c.Full = w
	return c, nil
}

// LoadConfigFull reads root/bonsai.yaml in full (ReadConfigFull). A missing file is an *Error whose
// errors.Is(err, fs.ErrNotExist) holds, as LoadConfig's is.
func LoadConfigFull(root string) (*Config, error) {
	raw, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		e := &Error{File: ConfigFile, Msg: "cannot be read: " + oneLine(err), Err: err,
			Next: "check the file's permissions, then run the command again"}
		if errors.Is(err, fs.ErrNotExist) {
			e.Msg = "is not in this checkout (" + filepath.ToSlash(root) + "), so it is not linked to Bonsai"
			e.Next = "run bonsai init in the project's main checkout"
		}
		return nil, e
	}
	return ReadConfigFull(raw)
}

// fileError gives a refusal of internal/format as this package's Error, naming the file.
func fileError(file string, err error) error {
	var re *format.ReadError
	if !errors.As(err, &re) {
		return &Error{File: file, Msg: oneLine(err), Next: "tell the person: Bonsai could not read the file", Err: err}
	}
	return &Error{File: file, Line: re.Line, Code: re.Code, Msg: re.Msg, Next: re.Next, Err: err}
}
