package selfcheck

// docs/install.md held to the install lines' one table (plan-5 note 5.7.3, 3): the page's first bash block under
// installHeads' Linux heading and its first powershell block under the Windows one, read as committed (LF), are the
// lines engine.ReleaseLines writes for 1.0.1 on amd64, byte for byte, and the installed-copy checks carry the table's
// origin words. A line changed on either side, here or in the table, fails here naming the page's line.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
)

// installDoc is the page, relative to the module root; installVersion the version its blocks show.
const (
	installDoc     = "docs/install.md"
	installVersion = "1.0.1"
)

// installHeads are the two headings whose first fenced block is a side's install lines, with that block's language.
var installHeads = []struct{ heading, lang, side string }{
	{"## Install on Linux and WSL", "bash", "linux"},
	{"## Install on Windows", "powershell", "windows"},
}

// installOrigin is the part of the install lines' origin check an installed copy's check repeats (with no tag: an
// installed copy's version is not fixed here).
var installOrigin = " -R " + engine.ReleaseRepo + " --signer-workflow " + engine.ReleaseWorkflow + " --deny-self-hosted-runners"

// installProblems returns one problem per way the page leaves the table, each "docs/install.md:<line>: what".
func installProblems(page string) []string {
	lines := engine.ReleaseLines(installVersion, "amd64")
	if lines == nil {
		return []string{"the table gives no lines for " + installVersion}
	}
	want := map[string][]string{"linux": lines.Linux, "windows": lines.Windows}
	text := strings.Split(page, "\n")
	var out []string
	for _, h := range installHeads {
		at := -1
		for i, l := range text {
			if l == h.heading {
				at = i
				break
			}
		}
		if at < 0 {
			out = append(out, fmt.Sprintf("%s: no heading %q", installDoc, h.heading))
			continue
		}
		open := -1
		for i := at + 1; i < len(text) && !strings.HasPrefix(text[i], "## "); i++ {
			if strings.HasPrefix(text[i], "```") {
				open = i
				break
			}
		}
		if open < 0 || text[open] != "```"+h.lang {
			out = append(out, fmt.Sprintf("%s:%d: the first block under %q is not a ```%s block", installDoc, at+1, h.heading, h.lang))
			continue
		}
		var got []string
		closed := false
		for i := open + 1; i < len(text); i++ {
			if text[i] == "```" {
				closed = true
				break
			}
			got = append(got, text[i])
		}
		if !closed {
			out = append(out, fmt.Sprintf("%s:%d: the %s block is never closed", installDoc, open+1, h.lang))
			continue
		}
		w := want[h.side]
		for i := 0; i < len(got) || i < len(w); i++ {
			switch {
			case i >= len(w):
				out = append(out, fmt.Sprintf("%s:%d: a line the table does not have: %s", installDoc, open+2+i, got[i]))
			case i >= len(got):
				out = append(out, fmt.Sprintf("%s:%d: the %s block lacks the table's line %d: %s", installDoc, open+2+len(got), h.side, i+1, w[i]))
			case got[i] != w[i]:
				out = append(out, fmt.Sprintf("%s:%d: the %s block's line %d is %q, the table's %q", installDoc, open+2+i, h.side, i+1, got[i], w[i]))
			}
		}
	}
	for _, place := range []string{"/usr/local/bin/bonsai", `'C:\Program Files\Bonsai\bonsai.exe'`} {
		line := "gh attestation verify " + place + installOrigin
		if !strings.Contains(page, "\n"+line+"\n") {
			out = append(out, fmt.Sprintf("%s: no installed-copy check, as a line of its own: %s", installDoc, line))
		}
	}
	return out
}

// TestInstallDocHeldToTheLines: the page as committed has none of the problems installProblems finds.
func TestInstallDocHeldToTheLines(t *testing.T) {
	for _, p := range installProblems(read(t, installDoc)) {
		t.Error(p)
	}
}

// TestInstallDocCatchesDrift: copies of the page, each changed one way, are each caught: a line changed on each side
// (the origin check's tag, a flag dropped), a line dropped, a line added, a block's language, a heading renamed, and an
// installed-copy check without --deny-self-hosted-runners.
func TestInstallDocCatchesDrift(t *testing.T) {
	page := read(t, installDoc)
	swap := func(old, new string) string {
		if strings.Count(page, old) < 1 {
			t.Fatalf("the page holds no %q to change", old)
		}
		return strings.Replace(page, old, new, 1)
	}
	cases := map[string]struct{ page, want string }{
		"the linux tag changed": {swap("linux_amd64.tar.gz -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1",
			"linux_amd64.tar.gz -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.2"), "the linux block's line 5"},
		"the windows check loosened": {swap("windows_amd64.zip -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1 --deny-self-hosted-runners",
			"windows_amd64.zip -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1"), "the windows block's line 5"},
		"the fingerprint line dropped": {swap("sha256sum -c --ignore-missing checksums.txt\n", ""), "the linux block's line 4"},
		"a line added":                 {swap("sh install.sh\n```", "sh install.sh\necho done\n```"), "a line the table does not have: echo done"},
		"a block's language":           {swap("```powershell\ncd (New-Item", "```ps1\ncd (New-Item"), "is not a ```powershell block"},
		"a heading renamed":            {swap("## Install on Linux and WSL\n", "## Install on Linux\n"), `no heading "## Install on Linux and WSL"`},
		"an installed-copy check loosened": {swap("gh attestation verify /usr/local/bin/bonsai -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --deny-self-hosted-runners",
			"gh attestation verify /usr/local/bin/bonsai -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml"), "no installed-copy check"},
	}
	for name, c := range cases {
		got := strings.Join(installProblems(c.page), "\n")
		if !strings.Contains(got, c.want) || !strings.Contains(got, installDoc) {
			t.Errorf("%s: want a problem naming %q, got:\n%s", name, c.want, got)
		}
	}
}
