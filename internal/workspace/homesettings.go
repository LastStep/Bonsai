package workspace

// This machine's settings, kept in the Bonsai home (spec section 4, section 6; contract section 3, section 10.1;
// plan-5 5.6.1): read and written by bonsai settings show and set, and read by the commands that use them.
//
//	<home>/workspaces/r-<16 hex>/settings.json   the workspace's (scope workspace): status_writes, status_command
//	<home>/settings.json                         the home's own (scope machine, --machine): cache_keep_days, and
//	                                              personal_from on Windows
//
// Each file is a JSON object of key to value, written by settings set only: two-space indent, LF line ends, ASCII,
// the keys in SettingKeys' order and then any key this Bonsai does not know, kept as read. A missing file holds every
// key's default. Example, a workspace a studio manages:
//
//	{
//	  "status_writes": "command",
//	  "status_command": "studio task move"
//	}
//
// and a home that cleans its pack cache:
//
//	{
//	  "cache_keep_days": 90
//	}
//
// SettingKeys is the keys' one home: settings show and set, their --help and the reference page of lists read it.
// The reader applies the same rules as the writer (CheckSetting), so a file edited by hand or forged never gives a
// command a value outside its rule: such a file is refused, naming the file and the key.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The two scopes of this machine's settings (bonsai.settings/1's scope).
const (
	ScopeWorkspace = "workspace" // the workspace's, in its machine folder
	ScopeMachine   = "machine"   // the Bonsai home's own, with --machine
)

// HomeSettingsFile is the home's own settings file, <home>/settings.json.
const HomeSettingsFile = "settings.json"

// SettingKey is one key of this machine's settings.
type SettingKey struct {
	Key     string // as settings set takes it
	Scope   string // ScopeWorkspace or ScopeMachine
	Windows bool   // read and set on Windows only
	Values  string // the values it takes, in words
	Default string // its default, in words
	Means   string // what it does
}

// SettingKeys are this machine's settings, in their order: the one home of the keys, their rules and defaults.
var SettingKeys = []SettingKey{
	{Key: "status_writes", Scope: ScopeWorkspace, Values: "agents or command", Default: "agents",
		Means: "who moves a task's status: agents, by editing its status line, or command, through status_command (a workspace a studio manages; contract section 10.1)"},
	{Key: "status_command", Scope: ScopeWorkspace,
		Values:  "with status_writes command only: one line of 1 to 100 characters, each a letter, a digit, a space or one of - _ . / :",
		Default: "none",
		Means:   "the command agents run to move a task's status, which the guard names in every refusal of a status edit in command mode"},
	{Key: "cache_keep_days", Scope: ScopeMachine, Values: "a whole number of days from 1 to 3650, or none", Default: "none",
		Means: "bonsai update --yes removes each pack clone in the home's cache/git/ unused for longer than this, never one the project's lock or bonsai.yaml names, and never an --adopt copy (spec section 6)"},
	{Key: "personal_from", Scope: ScopeMachine, Windows: true,
		Values:  "WSL's personal folder as Windows reaches it (such as //wsl.localhost/<distro>/home/<user>/.bonsai/personal), or none; never the home's own personal/ or a folder in it",
		Default: "none",
		Means:   "on Windows only: where the personal layer is read from, so one personal folder serves both sides (step 5.6.3)"},
}

// KeysOf lists the keys of a scope on an operating system ("windows", "linux"), in SettingKeys' order.
func KeysOf(scope, goos string) []SettingKey {
	var out []SettingKey
	for _, k := range SettingKeys {
		if k.Scope == scope && (!k.Windows || goos == "windows") {
			out = append(out, k)
		}
	}
	return out
}

// settingKey finds a key of any scope; nil when there is none.
func settingKey(name string) *SettingKey {
	for i := range SettingKeys {
		if SettingKeys[i].Key == name {
			return &SettingKeys[i]
		}
	}
	return nil
}

// The rule of status_command: one line of at most this many characters, from statusCommandChars.
const (
	StatusCommandMax   = 100
	statusCommandChars = "-_./: "
)

// StatusCommandOK reports whether a status_command keeps its rule: 1 to 100 characters, each an ASCII letter, a digit,
// a space or one of - _ . / :, and not spaces alone. The guard prints it in every command-mode refusal and agents run
// it as told, so it can hold no ;, $, quote, backtick or line break.
func StatusCommandOK(s string) bool {
	if len(s) == 0 || len(s) > StatusCommandMax || strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune(statusCommandChars, r)
		if !ok {
			return false
		}
	}
	return true
}

// The bounds of cache_keep_days.
const (
	CacheKeepMin = 1
	CacheKeepMax = 3650
)

// None is the value that sets a key back to having none (cache_keep_days, personal_from).
const None = "none"

// CheckSetting checks a value typed for a key (settings set's key=value) and gives it as the file stores it: a text,
// an int64 for cache_keep_days, or nil for none. home is the Bonsai home (personal_from may not name its personal/),
// goos the operating system. A value outside the rule gives a sentence saying why, naming the values the key takes.
func CheckSetting(key, value, home, goos string) (any, string) {
	k := settingKey(key)
	if k == nil {
		return nil, "is not a key of this machine's settings"
	}
	bad := func(why string) (any, string) {
		return nil, key + "=" + strconv.QuoteToASCII(value) + " " + why + ": " + key + " takes " + k.Values
	}
	switch key {
	case "status_writes":
		if value != StatusWritesAgents && value != StatusWritesCommand {
			return bad("is not one of its values")
		}
		return value, ""
	case "status_command":
		if !StatusCommandOK(value) {
			return bad("is outside its rule")
		}
		return value, ""
	case "cache_keep_days":
		if value == None {
			return nil, ""
		}
		n, ok := wholeDays(value)
		if !ok {
			return bad("is not a whole number of days from 1 to 3650")
		}
		return n, ""
	case "personal_from":
		if goos != "windows" {
			return nil, "personal_from is read on Windows only (where it points at WSL's personal folder): this machine is " + goos
		}
		if value == None {
			return nil, ""
		}
		if why := personalFromProblem(value, home); why != "" {
			return bad(why)
		}
		return winSlash(value), ""
	}
	return nil, "is not a key of this machine's settings"
}

// wholeDays reads cache_keep_days' number: digits only, 1 to 3650.
func wholeDays(s string) (int64, bool) {
	if s == "" || len(s) > 4 {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < CacheKeepMin || n > CacheKeepMax {
		return 0, false
	}
	return n, true
}

// personalFromProblem says why a personal_from path is outside its rule, or "": it must be an absolute Windows path
// (a drive's, X:/..., or a share's, //host/...), hold no control character, and be neither the home's own personal/
// nor a folder inside it (its own target, step 5.6.3).
func personalFromProblem(p, home string) string {
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return "holds a control character"
		}
	}
	s := winSlash(p)
	drive := len(s) >= 3 && ((s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z')) && s[1] == ':' && s[2] == '/'
	share := strings.HasPrefix(s, "//") && len(s) > 2 && s[2] != '/'
	if !drive && !share {
		return "is not an absolute Windows path (X:/... or //host/...)"
	}
	if home != "" {
		own := strings.ToLower(strings.TrimSuffix(winSlash(home), "/") + "/personal")
		low := strings.ToLower(strings.TrimSuffix(s, "/"))
		if low == own || strings.HasPrefix(low, own+"/") {
			return "names the home's own personal/ (or a folder in it), which is its target"
		}
	}
	return ""
}

// winSlash is a Windows path with forward slashes, on any system (filepath.ToSlash changes only this system's
// separator, and personal_from is a Windows path checked on Linux too, in the tests).
func winSlash(p string) string { return strings.ReplaceAll(p, `\`, "/") }

// storedProblem checks a value a settings file holds for a key, as the reader does: "" when it keeps the key's rule
// (null always does: the key's default), else why not.
func storedProblem(key string, v any, home, goos string) (any, string) {
	if v == nil {
		return nil, ""
	}
	switch key {
	case "cache_keep_days":
		if d, ok := v.(int64); ok && d >= CacheKeepMin && d <= CacheKeepMax {
			return d, "" // as Set stores it, before it is written
		}
		n, ok := v.(json.Number)
		if !ok {
			return nil, "cache_keep_days is " + schema.Show(v) + ", not a whole number of days from 1 to 3650"
		}
		d, ok := wholeDays(n.String())
		if !ok {
			return nil, "cache_keep_days is " + schema.Show(v) + ", not a whole number of days from 1 to 3650"
		}
		return d, ""
	default:
		s, ok := v.(string)
		if !ok {
			return nil, key + " is " + schema.Show(v) + ", not a text"
		}
		val, why := CheckSetting(key, s, home, goos)
		if why != "" {
			return nil, key + " is " + strconv.QuoteToASCII(s) + ", outside its rule (" + settingKey(key).Values + ")"
		}
		return val, ""
	}
}

// SettingInForce is one key with the value in force and where it came from (bonsai.settings/1's settings[]).
type SettingInForce struct {
	Key   string
	Value any    // a text, an int64, or nil for none
	From  string // file or default
}

// The values of SettingInForce.From.
const (
	FromFile    = "file"
	FromDefault = "default"
)

// SettingsFile is one of this machine's settings files, as read.
type SettingsFile struct {
	Scope  string
	Path   string        // absolute, the system's separators
	Exists bool          // the file is there
	Obj    schema.Object // its members as read; empty when it is missing
	home   string
	goos   string
}

// Show is the file's place for a person and for --json: absolute, forward slashes.
func (f *SettingsFile) Show() string { return filepath.ToSlash(f.Path) }

// OpenSettings reads the settings file of a scope: the workspace's (in the machine folder of main, the main checkout)
// or the home's own (main unused). A missing file is empty. Its error is an *Error with the Code read-failed, naming
// the file and the next step.
func OpenSettings(scope, home, main string) (*SettingsFile, error) {
	return openSettings(scope, home, main, runtime.GOOS)
}

func openSettings(scope, home, main, goos string) (*SettingsFile, error) {
	f := &SettingsFile{Scope: scope, home: home, goos: goos}
	if scope == ScopeMachine {
		f.Path = filepath.Join(home, HomeSettingsFile)
	} else {
		dir, err := MachineDir(home, main)
		if err != nil {
			return nil, err
		}
		f.Path = filepath.Join(dir, MachineSettingsFile)
	}
	raw, err := os.ReadFile(f.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return f, nil
	}
	if err != nil {
		return nil, &Error{File: f.Show(), Code: "read-failed", Msg: "cannot be read: " + oneLine(err), Err: err, Next: "check the file's permissions"}
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	o, isObject := v.(schema.Object)
	if err != nil || !isObject {
		return nil, &Error{File: f.Show(), Code: "read-failed", Msg: "is not a JSON object Bonsai reads", Next: f.fixNext()}
	}
	f.Exists, f.Obj = true, o
	return f, nil
}

// fixNext is the next step for a settings file that holds what Bonsai does not read.
func (f *SettingsFile) fixNext() string {
	return "a person fixes " + f.Show() + " (this machine's settings), or deletes it to go back to the defaults"
}

// InForce gives every key of the file's scope with the value in force, in SettingKeys' order, applying each key's
// rule to what the file holds: a value outside it is an *Error (Code read-failed) naming the file and the key.
// status_command is in force only with status_writes command (else none, its default).
func (f *SettingsFile) InForce() ([]SettingInForce, error) {
	return f.inForce(f.Obj)
}

func (f *SettingsFile) inForce(o schema.Object) ([]SettingInForce, error) {
	keys := KeysOf(f.Scope, f.goos)
	out := make([]SettingInForce, 0, len(keys))
	writes := StatusWritesAgents
	for _, k := range keys {
		raw, has := o.Get(k.Key)
		v, why := storedProblem(k.Key, raw, f.home, f.goos)
		if why != "" {
			return nil, &Error{File: f.Show(), Code: "read-failed", Msg: why, Next: f.fixNext()}
		}
		s := SettingInForce{Key: k.Key, Value: v, From: FromDefault}
		if has && v != nil {
			s.From = FromFile
		}
		switch k.Key {
		case "status_writes":
			if v == nil {
				s.Value = StatusWritesAgents
			}
			writes = s.Value.(string)
		case "status_command":
			if writes != StatusWritesCommand {
				s.Value, s.From = nil, FromDefault
			} else if v == nil {
				return nil, &Error{File: f.Show(), Code: "read-failed",
					Msg: "status_writes is command, but status_command names no command", Next: f.fixNext()}
			}
		}
		out = append(out, s)
	}
	return out, nil
}

// SettingError is a refusal of settings set's input (bad-value): nothing was written.
type SettingError struct{ Msg, Next string }

func (e *SettingError) Error() string { return e.Msg }

// Set checks every key=value first and then writes them all or none: an unknown key or a value outside its rule is
// a *SettingError and nothing is written. status_writes=command needs a status_command, given or already set;
// status_writes=agents drops status_command; none takes a key out of the file. The file is written byte-stable
// (SettingKeys' order, then the keys this Bonsai does not know, as read), staged and renamed with Windows' busy
// retries, and only when a value in force changes. It gives the keys whose value in force changed, in SettingKeys'
// order; a write that fails is an *Error with the Code write-failed.
func (f *SettingsFile) Set(pairs []string) ([]string, error) {
	keys := KeysOf(f.Scope, f.goos)
	var names []string
	for _, k := range keys {
		names = append(names, k.Key)
	}
	listNext := "give key=value with one of the " + f.Scope + "'s keys: " + strings.Join(names, ", ") + " (bonsai settings set --help)"
	if len(pairs) == 0 {
		return nil, &SettingError{Msg: "bonsai settings set needs at least one key=value", Next: listNext}
	}
	before, err := f.InForce()
	if err != nil {
		return nil, err
	}
	o := append(schema.Object{}, f.Obj...)
	given := map[string]bool{}
	for _, p := range pairs {
		key, value, ok := strings.Cut(p, "=")
		if !ok || key == "" {
			return nil, &SettingError{Msg: strconv.QuoteToASCII(p) + " is not key=value", Next: listNext}
		}
		k := settingKey(key)
		if k == nil || k.Scope != f.Scope {
			msg := strconv.QuoteToASCII(key) + " is not a key of the " + f.Scope + "'s settings"
			next := listNext
			if k != nil && k.Scope == ScopeMachine {
				next = "the home's own keys take --machine: bonsai settings set " + quotedPair(p) + " --machine"
			} else if k != nil {
				next = "the workspace's keys take no --machine: bonsai settings set " + quotedPair(p)
			}
			return nil, &SettingError{Msg: msg, Next: next}
		}
		if given[key] {
			return nil, &SettingError{Msg: key + " is given twice", Next: "give each key once"}
		}
		given[key] = true
		v, why := CheckSetting(key, value, f.home, f.goos)
		if why != "" {
			return nil, &SettingError{Msg: why, Next: "give one of the values " + key + " takes (bonsai settings set --help)"}
		}
		o = setMember(o, key, v)
	}
	if f.Scope == ScopeWorkspace {
		writes, _ := o.Get("status_writes")
		if writes == StatusWritesCommand {
			if c, _ := o.Get("status_command"); c == nil {
				return nil, &SettingError{Msg: "status_writes=command needs a status_command, and none is given or set",
					Next: `give both: bonsai settings set status_writes=command "status_command=<the command>"`}
			}
		} else {
			if given["status_command"] {
				return nil, &SettingError{Msg: "status_command is used with status_writes command only, and status_writes is agents",
					Next: `give both: bonsai settings set status_writes=command "status_command=<the command>"`}
			}
			o = setMember(o, "status_command", nil)
		}
	}
	after, err := f.inForce(o)
	if err != nil {
		return nil, err
	}
	var changed []string
	for i := range after {
		if !sameValue(before[i].Value, after[i].Value) {
			changed = append(changed, after[i].Key)
		}
	}
	if len(changed) == 0 {
		return []string{}, nil
	}
	raw, err := schema.Encode(ordered(o))
	if err != nil {
		return nil, &Error{File: f.Show(), Code: "write-failed", Msg: "cannot be written: " + oneLine(err), Err: err, Next: "report this to Bonsai's maintainers"}
	}
	if err := WriteFileAtomic(f.Path, raw); err != nil {
		return nil, &Error{File: f.Show(), Code: "write-failed", Msg: "cannot be written: " + oneLine(err), Err: err,
			Next: "check that the Bonsai home can be written (BONSAI_HOME), then run the command again"}
	}
	f.Exists, f.Obj = true, ordered(o)
	return changed, nil
}

// quotedPair quotes a key=value for a shell when it needs it, for a next step.
func quotedPair(s string) string {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("-_./:=", r)) {
			return `"` + strings.ReplaceAll(s, `"`, `'`) + `"`
		}
	}
	return s
}

// setMember sets a member of an object, keeping its place; nil takes it out.
func setMember(o schema.Object, key string, v any) schema.Object {
	i := o.Index(key)
	if v == nil {
		if i >= 0 {
			return append(o[:i:i], o[i+1:]...)
		}
		return o
	}
	if i >= 0 {
		o[i].Value = v
		return o
	}
	return append(o, schema.Member{Key: key, Value: v})
}

// ordered is an object with SettingKeys' keys first, in their order, then the others as they were.
func ordered(o schema.Object) schema.Object {
	out := schema.Object{}
	for _, k := range SettingKeys {
		if v, ok := o.Get(k.Key); ok {
			out = append(out, schema.Member{Key: k.Key, Value: v})
		}
	}
	for _, m := range o {
		if settingKey(m.Key) == nil {
			out = append(out, m)
		}
	}
	return out
}

// sameValue compares two values in force.
func sameValue(a, b any) bool { return schema.Equal(a, b) }

// HomeSettings is the home's own settings, as commands use them.
type HomeSettings struct {
	CacheKeepDays int    // 0 for none: the cache is never cleaned
	PersonalFrom  string // Windows only; "" for none
}

// LoadHomeSettings reads <home>/settings.json with its keys' rules (a missing file holds the defaults). Its error
// is an *Error naming the file and the key.
func LoadHomeSettings(home string) (HomeSettings, error) {
	var out HomeSettings
	f, err := OpenSettings(ScopeMachine, home, "")
	if err != nil {
		return out, err
	}
	in, err := f.InForce()
	if err != nil {
		return out, err
	}
	for _, s := range in {
		switch v := s.Value.(type) {
		case int64:
			if s.Key == "cache_keep_days" {
				out.CacheKeepDays = int(v)
			}
		case string:
			if s.Key == "personal_from" {
				out.PersonalFrom = v
			}
		}
	}
	return out, nil
}
