package stopgate

// A test project's commit, written by hand as git writes loose objects (zlib-compressed "<kind> <size>\0<body>", named
// by the SHA-1 of the uncompressed bytes), so the hand-made .git holds what the stop gate reads with git: bonsai.yaml as
// committed at HEAD (CommittedConfig). The commit holds bonsai.yaml alone, with a fixed author and time, so its id is
// known before any test runs (headSHA).

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// object is one loose object: its id and its bytes as git stores them before compression.
type object struct {
	id  string
	raw []byte
}

// objectOf frames a body as a git object of a kind.
func objectOf(kind string, body []byte) object {
	raw := append([]byte(fmt.Sprintf("%s %d\x00", kind, len(body))), body...)
	sum := sha1.Sum(raw)
	return object{id: hex.EncodeToString(sum[:]), raw: raw}
}

// commitObjects are the blob of bonsai.yaml holding yaml, a tree holding it, and a commit of that tree, in that order.
func commitObjects(yaml string) []object {
	blob := objectOf("blob", []byte(yaml))
	id, _ := hex.DecodeString(blob.id)
	tree := objectOf("tree", append([]byte("100644 bonsai.yaml\x00"), id...))
	commit := objectOf("commit", []byte("tree "+tree.id+"\nauthor t <t> 1760184000 +0000\ncommitter t <t> 1760184000 +0000\n\nthe fixture\n"))
	return []object{blob, tree, commit}
}

// commitOf is the id of the commit holding bonsai.yaml with yaml.
func commitOf(yaml string) string { return commitObjects(yaml)[2].id }

// headSHA is the commit the test project's HEAD names: bonsai.yaml as testYAML, committed.
var headSHA = commitOf(testYAML)

// commitYAML writes the commit holding bonsai.yaml with yaml into the git folder gitDir (with the files git needs to
// read it as a repository: config, objects/ and refs/), and gives its id. It moves no ref.
func commitYAML(t *testing.T, gitDir, yaml string) string {
	t.Helper()
	if _, err := os.Stat(filepath.Join(gitDir, "config")); err != nil {
		write(t, gitDir, "config", "[core]\n\trepositoryformatversion = 0\n\tbare = false\n")
	}
	for _, d := range []string{"refs/heads", "refs/tags"} {
		if err := os.MkdirAll(filepath.Join(gitDir, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	objs := commitObjects(yaml)
	for _, o := range objs {
		writeObject(t, gitDir, o)
	}
	return objs[2].id
}

// writeObject writes one loose object into the git folder gitDir, compressed as git compresses it.
func writeObject(t *testing.T, gitDir string, o object) {
	t.Helper()
	var b bytes.Buffer
	z := zlib.NewWriter(&b)
	if _, err := z.Write(o.raw); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	write(t, gitDir, "objects/"+o.id[:2]+"/"+o.id[2:], b.String())
}

// emptyCommit writes a commit of the empty tree into the git folder gitDir (which commitYAML has made a repository)
// and gives its id: a commit that holds no bonsai.yaml.
func emptyCommit(t *testing.T, gitDir string) string {
	t.Helper()
	tree := objectOf("tree", nil)
	commit := objectOf("commit", []byte("tree "+tree.id+"\nauthor t <t> 1760184000 +0000\ncommitter t <t> 1760184000 +0000\n\nempty\n"))
	writeObject(t, gitDir, tree)
	writeObject(t, gitDir, commit)
	return commit.id
}

// recommit commits bonsai.yaml with yaml in the main checkout root (the file and its commit), moves main's branch to
// it, and writes a green result proving it: the project as the gate first saw it, with another bonsai.yaml committed.
func recommit(t *testing.T, root, yaml string) string {
	t.Helper()
	write(t, root, "bonsai.yaml", yaml)
	sha := commitYAML(t, filepath.Join(root, ".git"), yaml)
	write(t, root, ".git/refs/heads/main", sha+"\n")
	writeResult(t, root, "T-0901", greenResult(sha))
	return sha
}
