package message

// Reading every message file (design/plan-5.md, note 5.5.10, 1): the main checkout's .bonsai/local/messages/, one
// folder per recipient address, one <id>.md file per message, each read by format.ReadMessage (Bonsai's own YAML
// rules, contract §2.4; an unknown field kept; a newer major refused as format too new). A file that does not read is
// reported in Malformed, with why and its thread where its frontmatter still names one, and skipped: one bad file
// never blocks the rest. Beside the format's rules, a file is malformed when:
//
//   - it is not a regular file (a folder or a link where a message file belongs), or a recipient's folder is a file
//     or a link, or cannot be listed;
//   - its name is not <id>.md, or is not its own id's;
//   - it is larger than MaxFile, so it is not read at all;
//   - it sits in another folder than its to's (builder@T-0901's is builder-T-0901), or its from or to names a folder
//     Windows keeps for a device (CheckAddress);
//   - its id is another file's too (the first by path is read, the next listed).
//
// A name starting with a dot is never a message (a write's staged file, .<id>.md.tmp-<random>, an editor's or a
// system's own file) and is passed over. Messages are sorted by created, then id; malformed files by path: never in
// map order, so every reader's output is byte-stable.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
)

// Entry is one message as read, with its file.
type Entry struct {
	Msg  *format.Message
	Path string // its file, project-relative in the main checkout, forward slashes
}

// Store is every message file under a main checkout's .bonsai/local/messages/, as read.
type Store struct {
	Messages  []Entry                   // the files that read, by created, then id
	Malformed []format.MessageMalformed // the files that do not, by path
}

// readDir, lstat and readFile are the reads Read makes, variables for the tests.
var (
	readDir  = os.ReadDir
	lstat    = os.Lstat
	readFile = os.ReadFile
)

// Read reads every message file of the main checkout at main. A missing messages folder is no message. Only a
// messages folder that is there and cannot be listed is an error: everything below it that does not read is listed
// in Malformed.
func Read(main string) (*Store, error) {
	st := &Store{Messages: []Entry{}, Malformed: []format.MessageMalformed{}}
	root := filepath.Join(main, filepath.FromSlash(Dir))
	folders, err := readDir(root)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return st, nil
	case err != nil:
		return nil, err
	}
	seen := map[string]string{} // an id to the path that read with it first
	for _, fe := range folders {
		name := fe.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		rel := Dir + "/" + name
		if fe.Type() != fs.ModeDir { // a file, a link, a junction or a device: anything but a plain folder
			st.malformed(rel, nil, "it is not a recipient's folder (a file or a link where only folders belong)")
			continue
		}
		files, err := readDir(filepath.Join(root, name))
		if err != nil {
			st.malformed(rel, nil, "the folder cannot be listed: "+oneLine(err))
			continue
		}
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
			path := rel + "/" + f.Name()
			msg, thread, why := readOne(filepath.Join(root, name, f.Name()), name, f.Name())
			if why == "" {
				if first, dup := seen[msg.ID]; dup {
					thread, why = &msg.Thread, "its id is also the id of "+safePath(first)+", read first"
				}
			}
			if why != "" {
				st.malformed(path, thread, why)
				continue
			}
			seen[msg.ID] = path
			st.Messages = append(st.Messages, Entry{Msg: msg, Path: path})
		}
	}
	sortEntries(st.Messages)
	sort.Slice(st.Malformed, func(i, j int) bool { return st.Malformed[i].Path < st.Malformed[j].Path })
	return st, nil
}

// readOne reads one file of the recipient folder folder: the message, or why it does not read and its thread where
// that still reads.
func readOne(path, folder, name string) (*format.Message, *string, string) {
	fi, err := lstat(path)
	switch {
	case err != nil:
		return nil, nil, "it cannot be read: " + oneLine(err)
	case !fi.Mode().IsRegular():
		return nil, nil, "it is not a regular file (a folder or a link where only message files belong)"
	case fi.Size() > MaxFile:
		return nil, nil, fmt.Sprintf("it is %d bytes, over the %d a message file may hold, so it is not read", fi.Size(), MaxFile)
	}
	if !strings.HasSuffix(name, ".md") || !IsID(strings.TrimSuffix(name, ".md")) {
		return nil, nil, "its name is not <id>.md, the name bonsai message gives a message's file"
	}
	raw, err := readFile(path)
	if err != nil {
		return nil, nil, "it cannot be read: " + oneLine(err)
	}
	return Parse(raw, folder, name)
}

// Parse reads one message file's bytes, held to bonsai.message/1 and to where it lies (its recipient folder and its
// file name): the message, or why it does not read and its thread where its frontmatter still names one.
func Parse(raw []byte, folder, name string) (*format.Message, *string, string) {
	msg, err := format.ReadMessage(raw)
	if err != nil {
		return nil, threadOf(raw), reasonOf(err)
	}
	thread := &msg.Thread
	switch {
	case name != msg.ID+".md":
		return nil, thread, "its id " + msg.ID + " is not its file's name"
	case CheckAddress(msg.From) != "":
		return nil, thread, "its from " + msg.From + " " + CheckAddress(msg.From)
	case CheckAddress(msg.To) != "":
		return nil, thread, "its to " + msg.To + " " + CheckAddress(msg.To)
	case FolderOf(msg.To) != folder:
		return nil, thread, "it is addressed to " + msg.To + ", whose folder is " + FolderOf(msg.To) + ", but it lies in " + ascii(folder)
	}
	if _, err := msg.Object(); err != nil {
		return nil, thread, reasonOf(err)
	}
	return msg, nil, ""
}

// reasonOf is a refusal of the format's reader as one ASCII sentence: where and what, never the next step (the
// malformed list's own next step is the file's).
func reasonOf(err error) string {
	var re *format.ReadError
	if !errors.As(err, &re) {
		return "it does not read as bonsai.message/1: " + oneLine(err)
	}
	where := "it does not read as " + re.Format
	if re.Line > 0 {
		where += fmt.Sprintf(" (line %d)", re.Line)
	}
	what := re.Msg
	if re.Field != "" {
		what = re.Field + " " + what
	}
	return where + ": " + ascii(what)
}

// threadOf gives the thread a file's frontmatter names, where its frontmatter still reads and its thread is an id.
func threadOf(raw []byte) *string {
	r := reader.ReadMarkdown(raw)
	if r.Outcome != reader.Accepted {
		return nil
	}
	v, _ := r.Value.Get("thread")
	if s, ok := v.(string); ok && IsID(s) {
		return &s
	}
	return nil
}

func (st *Store) malformed(path string, thread *string, why string) {
	st.Malformed = append(st.Malformed, format.MessageMalformed{Path: safePath(path), Thread: thread, Reason: why})
}

// safePath writes a path read from the disk as bonsai.messages/1 holds one: a colon, a backslash and a percent sign,
// which a name on Linux may hold and the format's paths may not, written as %3A, %5C and %25.
func safePath(p string) string {
	return strings.NewReplacer("%", "%25", ":", "%3A", `\`, "%5C").Replace(p)
}

// sortEntries sorts messages by created, then id (created is UTC to the millisecond in one layout, so its text sorts
// as its time).
func sortEntries(es []Entry) {
	sort.Slice(es, func(i, j int) bool {
		a, b := es[i].Msg, es[j].Msg
		if a.Created != b.Created {
			return a.Created < b.Created
		}
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		return es[i].Path < es[j].Path
	})
}

// Find gives the message whose id is id, or nil; when none reads, the malformed file named <id>.md, if there is one.
func (st *Store) Find(id string) (*Entry, *format.MessageMalformed) {
	for i := range st.Messages {
		if st.Messages[i].Msg.ID == id {
			return &st.Messages[i], nil
		}
	}
	for i := range st.Malformed {
		if strings.HasSuffix(st.Malformed[i].Path, "/"+id+".md") {
			return nil, &st.Malformed[i]
		}
	}
	return nil, nil
}
