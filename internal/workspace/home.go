package workspace

// The Bonsai home and a workspace's machine folder in it (contract §3), and the line-ending-blind file hash the
// lock records (contract §2.3, §14).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// HomeEnv names the variable that moves the Bonsai home; tests always set it to a temporary folder.
const HomeEnv = "BONSAI_HOME"

// Home returns the Bonsai home: BONSAI_HOME when it is set and not empty (made absolute), else .bonsai in the
// user's home folder (~/.bonsai on Linux, %USERPROFILE%\.bonsai on Windows). It creates nothing.
func Home() (string, error) {
	if h := os.Getenv(HomeEnv); h != "" {
		abs, err := filepath.Abs(h)
		if err != nil {
			return "", &Error{File: HomeEnv, Msg: "cannot be made absolute: " + oneLine(err), Err: err,
				Next: "set BONSAI_HOME to an absolute folder"}
		}
		return abs, nil
	}
	user, err := os.UserHomeDir()
	if err != nil || user == "" {
		if err == nil {
			err = errors.New("no home folder")
		}
		return "", &Error{File: "the Bonsai home", Msg: "cannot be found: " + oneLine(err), Err: err,
			Next: "set BONSAI_HOME to the folder Bonsai should use"}
	}
	return filepath.Join(user, ".bonsai"), nil
}

// MachineKey names a workspace's machine folder in the home, workspaces/<key>/: r- and the first 16 hex characters
// of the SHA-256 of the main checkout's real path (contract §3). The path hashed is the real path (symbolic links
// resolved), absolute, with forward slashes, and on Windows lower-cased, so one checkout always gives one key
// whatever letter case or separator a caller used.
func MachineKey(mainRoot string) (string, error) {
	abs, err := filepath.Abs(mainRoot)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return machineKey(real, runtime.GOOS), nil
}

func machineKey(real, goos string) string {
	p := filepath.ToSlash(real)
	if goos == "windows" {
		p = strings.ToLower(strings.ReplaceAll(p, `\`, "/"))
	}
	sum := sha256.Sum256([]byte(p))
	return "r-" + hex.EncodeToString(sum[:])[:16]
}

// HashLF returns the SHA-256 of a file's bytes after every CRLF is made LF, as 64 lower-case hex characters: the
// fingerprint the lock records for a file and a format-0 file (contract §2.3, §14), so a Windows checkout's line
// endings never read as an edit. A lone CR is kept: it is no line ending git makes.
func HashLF(raw []byte) string {
	sum := sha256.Sum256(bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n")))
	return hex.EncodeToString(sum[:])
}
