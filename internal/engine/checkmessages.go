package engine

// check's warning on the messages between agents (set 8; plan-5 5.5.10 note 5), the one word of
// format.MessageCheckWords, their one home: message-malformed, a file under the main checkout's
// .bonsai/local/messages/ that does not read as bonsai.message/1 (internal/message's reader, the one every command
// uses, names why). A warning, never a finding: one bad file never changes the exit code nor blocks the rest, and
// bonsai messages lists it as malformed beside the threads. The messages are found where every writer of
// .bonsai/local/ writes (workspace.FindLocal: from a worktree, its main checkout's).

import (
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// checkMessages adds a warning for each message file that does not read, and one for a messages folder that cannot
// be listed.
func (r *CheckResult) checkMessages() {
	if r.Config == nil {
		return
	}
	main := workspace.FindLocal(r.Root, r.Config.ID).Main
	st, err := message.Read(main)
	if err != nil {
		r.add("message-malformed", message.Dir, "", message.Dir+"/ cannot be listed, so no message in it was read: "+oneLine(err.Error()),
			"check that "+message.Dir+"/ can be read, then run: bonsai check")
		return
	}
	for _, m := range st.Malformed {
		r.add("message-malformed", m.Path, "", m.Path+" does not read as a message between agents: "+m.Reason,
			"leave the file as it is (only bonsai message writes one; bonsai messages lists it as malformed); if its words "+
				"are needed, its writer writes them again with bonsai message (bonsai check --schema bonsai.message lists every field)")
	}
}
