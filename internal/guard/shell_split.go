package guard

// The delete check's word splitter (plan-5 5.3.3 note 1). A command line is cut into simple commands, at ;, &&,
// ||, |, & and line ends, and each into words by the quoting of the shell that runs it: bash's (single and double
// quotes, the backslash, $'...'), PowerShell's (single and double quotes, the backtick, here-strings, the curly
// quotes and dashes PowerShell reads as plain ones) and cmd.exe's (double quotes, the caret). The text inside $( ),
// backticks, ( ), { }, <( ) and >( ), and PowerShell's $( ), @( ), @{ } and blocks, is split again where it stands,
// to a depth of three; the judge (shell.go) splits again the text a command gives a shell (bash -c, eval, cmd /c,
// powershell -Command, a here-document fed to bash), counting the same depth.
//
// It is a splitter, not a shell reader (question A, spec §19): nothing is expanded or run. A variable or a
// substitution stays as written and marks its word (expand), a glob character marks its word (glob) whether quoted
// or not (git and PowerShell's -Path read a quoted one as a glob too), and bash's brace lists ({a,b}) are written
// out, since they only spell out names. A here-document's body is kept for the judge, which splits it only when it
// is fed to a shell, and its substitutions are split where they stand unless its delimiter is quoted. A line it
// cannot split (a quote or bracket left open, nesting past three) gives a reason instead of a list; the judge then
// refuses the line only when it holds a delete word.
//
// Every index is checked before it is read, and every loop moves forward, so no input makes it panic or loop
// (FuzzShellDecide holds it); its time is linear in the line, each byte read at most once per level of nesting.

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// shDialect is the shell whose quoting a text is split by.
type shDialect uint8

const (
	shBash  shDialect = iota // bash: the Bash and Monitor tools (Git Bash on Windows), and wsl's command
	shPower                  // PowerShell: the PowerShell tool, and the text given to powershell or pwsh
	shCmd                    // cmd.exe: the text given to cmd /c
)

// shMaxDepth is how many levels of nested text are split again (plan-5 5.3.3 note 1: "to a depth of three").
const shMaxDepth = 3

// shMaxRecursion caps the splitter's own recursion inside one level of nesting (quotes within ${...} within
// quotes), so no input can grow the stack without bound.
const shMaxRecursion = 64

// shMaxBraceWords caps the words one bash brace list writes out; a word that would write out more counts as
// unnamed (expand), as a variable does.
const shMaxBraceWords = 64

// shWord is one word of a simple command.
type shWord struct {
	text   string   // with quotes taken off and escapes read; a variable or substitution kept as written
	raw    string   // as written in the line
	quoted bool     // a part of it was quoted or escaped
	glob   bool     // it holds a glob character (* ? and, but in cmd, [), quoted or not, or an extglob
	expand bool     // it holds a variable, a substitution, ${...}, %NAME%, !NAME!, a PowerShell block or splat
	assign bool     // bash: a NAME=value word
	param  bool     // PowerShell: an unquoted parameter name (-Name, -Name:value; any dash PowerShell reads)
	comma  bool     // PowerShell: joined to the next word by an unquoted comma, in one array argument
	splat  bool     // PowerShell: @name, a splat whose parameters cannot be read
	item   bool     // find's {} in the command -exec runs: an item find gives (set by the judge, not the splitter)
	alts   []string // bash: the words its brace lists write out, empty ones dropped; nil when it has none
}

// shFeed is text given to a command's standard input: a here-document or a here-string.
type shFeed struct {
	text   string
	nested []shNest // the substitutions in an unquoted here-document's body, which run whatever reads it
}

// shNest is a list split from text nested in a command: a substitution, a group, a block.
type shNest struct {
	list *shList
	sub  bool // it runs apart (a subshell, a substitution), so a cd inside it stays there
}

// shCmd is one simple command of a list, or a group.
type shCommand struct {
	words  []shWord
	group  *shList   // ( ) or { } at the start of a bash command; nil for a simple command
	sub    bool      // the group is a subshell
	op     string    // what follows it: ";", "&&", "||", "|", "&", or "" at the end
	piped  bool      // fed by a pipe from the command before it
	nested []shNest  // substitutions, process substitutions, groups and blocks inside its words
	feeds  []*shFeed // its here-documents and here-strings
}

// shList is a command list.
type shList struct {
	cmds  []shCommand
	depth int // 0 for the line itself, one more for each level of nesting
}

// shSplit splits text read by dialect d at nesting depth depth. It gives the list, or why the text cannot be split.
func shSplit(text string, d shDialect, depth int) (*shList, string) {
	if depth > shMaxDepth {
		return nil, "it nests more than three levels deep"
	}
	sc := &shScan{s: text, d: d, depth: depth}
	l := sc.list(0)
	if sc.err == "" && sc.i < len(sc.s) {
		sc.fail("a closing bracket has no opening one")
	}
	sc.readDocs()
	if sc.err != "" {
		return nil, sc.err
	}
	return l, ""
}

// shScan is one splitting of a text.
type shScan struct {
	s     string
	i     int
	d     shDialect
	depth int
	rec   int
	err   string
	docs  []shDoc // here-documents waiting for the end of their line
}

// shDoc is a here-document whose body starts after the current line.
type shDoc struct {
	feed   *shFeed
	delim  string
	strip  bool // <<- : leading tabs are dropped from each line
	quoted bool // a quoted delimiter: the body is not expanded
}

func (sc *shScan) fail(why string) {
	if sc.err == "" {
		sc.err = why
	}
}

// enter counts one level of the splitter's own recursion; leave undoes it.
func (sc *shScan) enter() bool {
	sc.rec++
	if sc.rec > shMaxRecursion {
		sc.fail("it nests quotes and brackets too deeply")
		return false
	}
	return true
}

func (sc *shScan) leave() { sc.rec-- }

func (sc *shScan) at(p string) bool { return strings.HasPrefix(sc.s[sc.i:], p) }

// adv moves on n bytes, never past the end.
func (sc *shScan) adv(n int) {
	sc.i += n
	if sc.i > len(sc.s) {
		sc.i = len(sc.s)
	}
}

// list reads commands until the end of the text or the closer that ends a nested list: ')' or '}' (0 for none).
// It stops on the closer without taking it.
func (sc *shScan) list(closer byte) *shList {
	l := &shList{depth: sc.depth}
	piped := false
	for sc.err == "" {
		start := sc.i
		sc.skip(true)
		if sc.i >= len(sc.s) || sc.closes(closer) {
			break
		}
		if c := sc.s[sc.i]; c == ')' || c == '}' && sc.d != shBash {
			// A closer with no opener at this level (a case pattern's ")"): a separator, as far as the check cares.
			sc.i++
			piped = false
			continue
		}
		cmd := sc.command(closer)
		cmd.piped = piped
		cmd.op = sc.operator()
		piped = cmd.op == "|"
		if len(cmd.words) > 0 || cmd.group != nil || len(cmd.nested) > 0 || len(cmd.feeds) > 0 {
			l.cmds = append(l.cmds, cmd)
		}
		if sc.i == start && sc.err == "" {
			sc.i++ // nothing read: never loop
		}
	}
	return l
}

// closes reports whether the scanner stands on the closer of the list it reads.
func (sc *shScan) closes(closer byte) bool {
	if closer == 0 || sc.i >= len(sc.s) || sc.s[sc.i] != closer {
		return false
	}
	if closer == '}' && sc.d == shBash {
		return sc.wordEnds(sc.i+1, " \t\r\n;&|)<>")
	}
	return true
}

// wordEnds reports whether position j ends a word: the end of the text, or one of the bytes in ends.
func (sc *shScan) wordEnds(j int, ends string) bool {
	return j >= len(sc.s) || strings.IndexByte(ends, sc.s[j]) >= 0
}

// blank gives the length of the blank at position i, 0 for none: spaces and tabs, a carriage return, a line
// continuation; PowerShell's Unicode spaces; cmd.exe's commas and semicolons.
func (sc *shScan) blank(i int) int {
	s := sc.s
	if i >= len(s) {
		return 0
	}
	switch c := s[i]; c {
	case ' ', '\t', '\r', '\f', '\v':
		return 1
	case '\\':
		if sc.d == shBash && i+1 < len(s) && s[i+1] == '\n' {
			return 2
		}
	case '`':
		if sc.d == shPower && i+1 < len(s) && (s[i+1] == '\n' || s[i+1] == '\r' && i+2 < len(s) && s[i+2] == '\n') {
			if s[i+1] == '\r' {
				return 3
			}
			return 2
		}
	case '^':
		if sc.d == shCmd && i+1 < len(s) && s[i+1] == '\n' {
			return 2
		}
	case ',', ';':
		if sc.d == shCmd {
			return 1
		}
	default:
		if sc.d == shPower && c >= utf8.RuneSelf {
			r, n := utf8.DecodeRuneInString(s[i:])
			if r != utf8.RuneError && unicode.IsSpace(r) {
				return n
			}
		}
	}
	return 0
}

// skip passes blanks and comments, and line ends when lines is set (reading the here-documents each one ends).
func (sc *shScan) skip(lines bool) {
	for sc.i < len(sc.s) && sc.err == "" {
		if n := sc.blank(sc.i); n > 0 {
			sc.i += n
			continue
		}
		c := sc.s[sc.i]
		switch {
		case c == '\n' && lines:
			sc.i++
			sc.readDocs()
		case c == '#' && sc.d != shCmd:
			for sc.i < len(sc.s) && sc.s[sc.i] != '\n' {
				sc.i++
			}
		case sc.d == shPower && sc.at("<#"):
			end := strings.Index(sc.s[sc.i+2:], "#>")
			if end < 0 {
				sc.fail("a block comment left open")
				return
			}
			sc.i += end + 4
		case sc.d == shCmd && sc.at("::"):
			for sc.i < len(sc.s) && sc.s[sc.i] != '\n' {
				sc.i++
			}
		default:
			return
		}
	}
}

// operator reads what ends a command: ;, &&, ||, |, & or a line end (given as ";"). It gives "" at a closer or the
// end of the text.
func (sc *shScan) operator() string {
	sc.skip(false)
	if sc.i >= len(sc.s) || sc.err != "" {
		return ""
	}
	s, i := sc.s, sc.i
	two := ""
	if i+1 < len(s) {
		two = s[i : i+2]
	}
	switch {
	case two == "&&" || two == "||":
		sc.i += 2
		return two
	case sc.d == shBash && (two == ";;" || two == ";&"):
		sc.i += 2
		if sc.i < len(s) && s[sc.i] == '&' {
			sc.i++
		}
		return ";"
	case sc.d == shBash && two == "|&":
		sc.i += 2
		return "|"
	case s[i] == ';' || s[i] == '&':
		sc.i++
		return string(s[i])
	case s[i] == '|':
		sc.i++
		return "|"
	case s[i] == '\n':
		sc.i++
		sc.readDocs()
		return ";"
	}
	return ""
}

// command reads one command: a bash group, or a simple command's words and redirections.
func (sc *shScan) command(closer byte) shCommand {
	var cmd shCommand
	if sc.d == shBash && sc.i < len(sc.s) {
		switch {
		case sc.at("(("):
			sc.i += 2
			sc.span('(', ')', 2, &cmd) // an arithmetic command: no words, but its substitutions run
		case sc.at("("):
			sc.i++
			cmd.group = sc.nested(')')
			cmd.sub = true
		case sc.at("{") && sc.wordEnds(sc.i+1, " \t\r\n"):
			sc.i++
			cmd.group = sc.nested('}')
		}
	}
	sc.words(&cmd, closer)
	return cmd
}

// nested reads a list up to its closer and takes the closer, one level deeper.
func (sc *shScan) nested(closer byte) *shList {
	if sc.depth >= shMaxDepth {
		sc.fail("it nests more than three levels deep")
		return nil
	}
	sc.depth++
	l := sc.list(closer)
	sc.depth--
	if sc.err != "" {
		return nil
	}
	if sc.i >= len(sc.s) || sc.s[sc.i] != closer {
		sc.fail("a bracket left open")
		return nil
	}
	sc.i++
	return l
}

// words reads a simple command's words and redirections up to what ends it.
func (sc *shScan) words(cmd *shCommand, closer byte) {
	for sc.err == "" {
		start := sc.i
		sc.skip(false)
		if sc.i >= len(sc.s) {
			return
		}
		if sc.redirect(cmd) {
			continue
		}
		if sc.ends(cmd, closer) {
			return
		}
		c := sc.s[sc.i]
		switch {
		case sc.d == shBash && c == '(':
			// A ( inside a command (f () { ... }, a stray group): its text is split as a subshell's.
			sc.i++
			if l := sc.nested(')'); l != nil {
				cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
			}
		case sc.d == shBash && c == '{' && sc.wordEnds(sc.i+1, " \t\r\n"):
			// A brace group inside a command (a function's body after f () or function f).
			sc.i++
			if l := sc.nested('}'); l != nil {
				cmd.nested = append(cmd.nested, shNest{list: l})
			}
		case sc.d != shBash && (c == '(' || c == '{' && sc.d == shPower):
			p := sc.i
			sc.i++
			closeBy := byte(')')
			if c == '{' {
				closeBy = '}'
			}
			if l := sc.nested(closeBy); l != nil {
				cmd.nested = append(cmd.nested, shNest{list: l, sub: c == '('})
				cmd.words = append(cmd.words, shWord{text: sc.s[p:sc.i], raw: sc.s[p:sc.i], expand: true})
			}
		case sc.d == shPower && c == '&' && len(cmd.words) == 0 && !sc.at("&&"):
			sc.i++ // PowerShell's call operator: the next word is the program
			cmd.words = append(cmd.words, shWord{text: "&", raw: "&"})
		case sc.d == shPower && sc.at("--%") && sc.wordEnds(sc.i+3, " \t\r\n"):
			sc.stopParsing(cmd)
		default:
			w := sc.word(cmd)
			if sc.d == shPower && w.comma && w.text == "" && len(cmd.words) > 0 {
				cmd.words[len(cmd.words)-1].comma = true // a, b written with a space before the comma
				break
			}
			cmd.words = append(cmd.words, w)
		}
		if sc.i == start && sc.err == "" {
			sc.i++ // nothing read: never loop
		}
	}
}

// ends reports whether the scanner stands on what ends the command: an operator, a line end or a closer.
func (sc *shScan) ends(cmd *shCommand, closer byte) bool {
	s, i := sc.s, sc.i
	c := s[i]
	switch c {
	case '\n', ';', '|', ')':
		return true
	case '&':
		if sc.d == shPower && len(cmd.words) == 0 && !sc.at("&&") {
			return false // the call operator
		}
		return true
	case '}':
		if sc.d == shPower {
			return true
		}
		return sc.d == shBash && closer == '}' && sc.wordEnds(i+1, " \t\r\n;&|)<>")
	}
	return false
}

// redirect reads a redirection and drops its target (the target's substitutions still count). It gives false
// when the scanner does not stand on one.
func (sc *shScan) redirect(cmd *shCommand) bool {
	s, i := sc.s, sc.i
	j := i
	if sc.d == shBash || sc.d == shCmd || sc.d == shPower {
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if sc.d == shPower && j == i && j < len(s) && s[j] == '*' {
			j++
		}
	}
	if j < len(s) && j == i && sc.d == shBash && s[j] == '&' && j+1 < len(s) && s[j+1] == '>' {
		j++ // &> and &>>
	}
	if j >= len(s) || s[j] != '<' && s[j] != '>' {
		return false
	}
	if j+1 < len(s) && s[j+1] == '(' && sc.d == shBash {
		return false // <( ) and >( ): a process substitution, read as a word
	}
	if sc.d == shBash && s[j] == '<' && strings.HasPrefix(s[j:], "<<<") {
		sc.i = j + 3
		sc.skip(false)
		w := sc.word(cmd)
		cmd.feeds = append(cmd.feeds, &shFeed{text: w.text})
		return true
	}
	if sc.d == shBash && s[j] == '<' && strings.HasPrefix(s[j:], "<<") {
		strip := strings.HasPrefix(s[j:], "<<-")
		sc.i = j + 2
		if strip {
			sc.i++
		}
		sc.skip(false)
		w := sc.word(cmd)
		feed := &shFeed{}
		cmd.feeds = append(cmd.feeds, feed)
		sc.docs = append(sc.docs, shDoc{feed: feed, delim: w.text, strip: strip, quoted: w.quoted})
		return true
	}
	k := j + 1
	for k < len(s) && (s[k] == '>' || s[k] == '<' || s[k] == '&' || s[k] == '|') && k < j+3 {
		k++
	}
	sc.i = k
	if sc.d == shPower && k < len(s) && s[k-1] == '&' {
		// 2>&1 and *>&1: the stream number is the target
		for sc.i < len(s) && s[sc.i] >= '0' && s[sc.i] <= '9' {
			sc.i++
		}
		return true
	}
	sc.skip(false)
	if sc.i < len(s) && !sc.ends(cmd, 0) {
		_ = sc.word(cmd)
	}
	return true
}

// readDocs reads the bodies of the here-documents waiting for this line's end, from the scanner's place.
func (sc *shScan) readDocs() {
	docs := sc.docs
	sc.docs = nil
	for _, d := range docs {
		var body strings.Builder
		for sc.i < len(sc.s) {
			line := sc.s[sc.i:]
			next := len(sc.s)
			if k := strings.IndexByte(line, '\n'); k >= 0 {
				line = line[:k]
				next = sc.i + k + 1
			}
			sc.i = next
			if d.strip {
				line = strings.TrimLeft(line, "\t")
			}
			if line == d.delim {
				break
			}
			body.WriteString(line)
			body.WriteByte('\n')
		}
		d.feed.text = body.String()
		if !d.quoted && sc.d == shBash {
			d.feed.nested = sc.expansions(d.feed.text)
		}
	}
}

// expansions splits the substitutions in a here-document's body, $( ), backticks and ${ }, each where it stands.
func (sc *shScan) expansions(text string) []shNest {
	sub := &shScan{s: text, d: sc.d, depth: sc.depth}
	var holder shCommand
	for sub.i < len(sub.s) && sub.err == "" {
		switch c := sub.s[sub.i]; {
		case c == '\\':
			sub.adv(2)
		case c == '$':
			var wb shBuild
			sub.bashDollar(&wb, &holder, true)
		case c == '`':
			var wb shBuild
			sub.backquote(&wb, &holder)
		default:
			sub.i++
		}
	}
	if sub.err != "" {
		sc.fail(sub.err)
	}
	return holder.nested
}

// stopParsing reads PowerShell's --% : the rest of the line goes to the program as written, split at spaces.
func (sc *shScan) stopParsing(cmd *shCommand) {
	sc.i += 3
	end := strings.IndexByte(sc.s[sc.i:], '\n')
	rest := sc.s[sc.i:]
	if end >= 0 {
		rest = rest[:end]
	}
	sc.i += len(rest)
	for _, f := range strings.Fields(rest) {
		w := shWord{text: f, raw: f}
		w.glob = strings.ContainsAny(f, "*?[")
		w.expand = strings.Count(f, "%") >= 2
		cmd.words = append(cmd.words, w)
	}
}

// shBuild builds one word.
type shBuild struct {
	b   []byte
	unq []bool // per byte: written outside quotes, so a bash brace there may write out a list
	w   shWord
}

// lit adds literal text, marking a glob character.
func (wb *shBuild) lit(s string, quoted bool, globs string) {
	for k := 0; k < len(s); k++ {
		c := s[k]
		if strings.IndexByte(globs, c) >= 0 {
			wb.w.glob = true
		}
		wb.b = append(wb.b, c)
		wb.unq = append(wb.unq, !quoted)
	}
	if quoted {
		wb.w.quoted = true
	}
}

// exp adds an expansion as written: the word can no longer be named.
func (wb *shBuild) exp(raw string) {
	wb.w.expand = true
	wb.b = append(wb.b, raw...)
	for k := 0; k < len(raw); k++ {
		wb.unq = append(wb.unq, false)
	}
}

// word reads one word in the scanner's dialect.
func (sc *shScan) word(cmd *shCommand) shWord {
	if !sc.enter() {
		return shWord{}
	}
	defer sc.leave()
	switch sc.d {
	case shPower:
		return sc.psWord(cmd)
	case shCmd:
		return sc.cmdWord(cmd)
	}
	return sc.bashWord(cmd)
}

// bashWord reads one word by bash's quoting.
func (sc *shScan) bashWord(cmd *shCommand) shWord {
	const globs = "*?["
	start := sc.i
	var wb shBuild
	s := sc.s
loop:
	for sc.i < len(s) && sc.err == "" {
		c := s[sc.i]
		switch c {
		case ' ', '\t', '\r', '\n', ';', '&', '|', ')':
			break loop
		case '<', '>':
			if sc.i+1 < len(s) && s[sc.i+1] == '(' {
				p := sc.i
				sc.i += 2
				if l := sc.nested(')'); l != nil {
					cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
				}
				wb.exp(s[p:sc.i])
				continue
			}
			break loop
		case '(':
			n := len(wb.b)
			switch {
			case n > 0 && wb.unq[n-1] && strings.IndexByte("?*+@!", wb.b[n-1]) >= 0:
				p := sc.i // an extglob: ?( ), *( ), +( ), @( ), !( )
				sc.i++
				sc.span('(', ')', 1, cmd)
				wb.lit(s[p:sc.i], false, "")
				wb.w.glob = true
			case n > 0 && wb.unq[n-1] && wb.b[n-1] == '=':
				p := sc.i // an array: a=( ... )
				sc.i++
				if l := sc.nested(')'); l != nil {
					cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
				}
				wb.exp(s[p:sc.i])
			case n > 0 && sc.i+1 < len(s) && s[sc.i+1] == ')':
				sc.i += 2 // f(): a function's name
				break loop
			default:
				break loop
			}
		case '\'':
			end := strings.IndexByte(s[sc.i+1:], '\'')
			if end < 0 {
				sc.fail("a quote left open")
				break loop
			}
			wb.lit(s[sc.i+1:sc.i+1+end], true, globs)
			sc.i += end + 2
		case '"':
			sc.i++
			sc.bashDouble(&wb, cmd)
		case '\\':
			if sc.i+1 >= len(s) {
				wb.lit(`\`, false, "")
				sc.i++
				continue
			}
			if s[sc.i+1] == '\n' {
				sc.i += 2
				continue
			}
			wb.lit(s[sc.i+1:sc.i+2], true, globs)
			sc.i += 2
		case '$':
			sc.bashDollar(&wb, cmd, false)
		case '`':
			sc.backquote(&wb, cmd)
		default:
			wb.lit(s[sc.i:sc.i+1], false, globs)
			sc.i++
		}
	}
	wb.w.raw = s[start:sc.i]
	wb.w.text = string(wb.b)
	wb.w.assign = shIsAssign(wb.b, wb.unq)
	if !wb.w.assign {
		alts, ok, seq := shBraces(wb.b, wb.unq)
		switch {
		case !ok:
			wb.w.expand = true
		case seq:
			wb.w.glob = true
		case alts != nil:
			wb.w.alts = alts
		}
	}
	return wb.w
}

// bashDouble reads the inside of a double-quoted string, after its opening quote.
func (sc *shScan) bashDouble(wb *shBuild, cmd *shCommand) {
	if !sc.enter() {
		return
	}
	defer sc.leave()
	const globs = "*?["
	s := sc.s
	wb.w.quoted = true
	for sc.i < len(s) && sc.err == "" {
		switch c := s[sc.i]; c {
		case '"':
			sc.i++
			return
		case '\\':
			if sc.i+1 < len(s) {
				switch n := s[sc.i+1]; n {
				case '\n':
					sc.i += 2
					continue
				case '$', '`', '"', '\\':
					wb.lit(string(n), true, globs)
					sc.i += 2
					continue
				}
			}
			wb.lit(`\`, true, "")
			sc.i++
		case '$':
			sc.bashDollar(wb, cmd, true)
		case '`':
			sc.backquote(wb, cmd)
		default:
			wb.lit(s[sc.i:sc.i+1], true, globs)
			sc.i++
		}
	}
	sc.fail("a quote left open")
}

// bashDollar reads what starts with $: $'...', $"...", $(( )), $( ), ${ }, $[ ], a variable, or a plain $.
func (sc *shScan) bashDollar(wb *shBuild, cmd *shCommand, inDouble bool) {
	if !sc.enter() {
		return
	}
	defer sc.leave()
	s, i := sc.s, sc.i
	if i+1 >= len(s) {
		wb.lit("$", inDouble, "")
		sc.i++
		return
	}
	switch n := s[i+1]; {
	case n == '\'' && !inDouble:
		text, end, ok := shANSIC(s, i+2)
		if !ok {
			sc.fail("a quote left open")
			sc.i = len(s)
			return
		}
		wb.lit(text, true, "*?[")
		sc.i = end
	case n == '"' && !inDouble:
		sc.i += 2
		sc.bashDouble(wb, cmd)
	case n == '(' && i+2 < len(s) && s[i+2] == '(':
		sc.i += 3
		sc.span('(', ')', 2, cmd)
		wb.exp(s[i:sc.i])
	case n == '(':
		sc.i += 2
		if l := sc.nested(')'); l != nil {
			cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
		}
		wb.exp(s[i:sc.i])
	case n == '{':
		sc.i += 2
		sc.span('{', '}', 1, cmd)
		wb.exp(s[i:sc.i])
	case n == '[':
		sc.i += 2
		sc.span('[', ']', 1, cmd)
		wb.exp(s[i:sc.i])
	case shNameStart(n):
		j := i + 2
		for j < len(s) && shNameChar(s[j]) {
			j++
		}
		wb.exp(s[i:j])
		sc.i = j
	case n >= '0' && n <= '9' || strings.IndexByte("@*#?$!-", n) >= 0:
		wb.exp(s[i : i+2])
		sc.i = i + 2
	default:
		wb.lit("$", inDouble, "")
		sc.i++
	}
}

// backquote reads a `...` substitution and splits its text one level deeper.
func (sc *shScan) backquote(wb *shBuild, cmd *shCommand) {
	s, i := sc.s, sc.i
	j := i + 1
	var inner []byte
	for j < len(s) {
		c := s[j]
		if c == '\\' && j+1 < len(s) && (s[j+1] == '`' || s[j+1] == '\\' || s[j+1] == '$') {
			inner = append(inner, s[j+1])
			j += 2
			continue
		}
		if c == '`' {
			break
		}
		inner = append(inner, c)
		j++
	}
	if j >= len(s) {
		sc.fail("a backquote left open")
		sc.i = len(s)
		return
	}
	sc.i = j + 1
	wb.exp(s[i:sc.i])
	l, why := shSplit(string(inner), sc.d, sc.depth+1)
	if why != "" {
		sc.fail(why)
		return
	}
	cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
}

// span passes text up to the closer that balances need openers already read ($(( )), (( )), ${ }, $[ ], an
// extglob), splitting the substitutions inside it where they stand.
func (sc *shScan) span(open, closeBy byte, need int, cmd *shCommand) {
	if !sc.enter() {
		return
	}
	defer sc.leave()
	s := sc.s
	depth := need
	for sc.i < len(s) && sc.err == "" {
		c := s[sc.i]
		switch {
		case c == '\\':
			sc.adv(2)
		case c == '\'' && open == '{':
			end := strings.IndexByte(s[sc.i+1:], '\'')
			if end < 0 {
				sc.fail("a quote left open")
				return
			}
			sc.i += end + 2
		case c == '"':
			sc.i++
			var wb shBuild
			sc.bashDouble(&wb, cmd)
		case c == '`':
			var wb shBuild
			sc.backquote(&wb, cmd)
		case c == '$' && sc.i+1 < len(s) && s[sc.i+1] == '(' && !(sc.i+2 < len(s) && s[sc.i+2] == '('):
			sc.i += 2
			if l := sc.nested(')'); l != nil {
				cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
			}
		case c == open:
			depth++
			sc.i++
		case c == closeBy:
			depth--
			sc.i++
			if depth == 0 {
				return
			}
		default:
			sc.i++
		}
	}
	sc.fail("a bracket left open")
}

// shANSIC reads bash's $'...' from position i, after the quote: the text with its escapes read, and the position
// after the closing quote.
func shANSIC(s string, i int) (string, int, bool) {
	var b []byte
	hex := func(max int) (rune, int) {
		var v rune
		k := 0
		for ; k < max && i+k < len(s); k++ {
			c := s[i+k]
			switch {
			case c >= '0' && c <= '9':
				v = v*16 + rune(c-'0')
			case c >= 'a' && c <= 'f':
				v = v*16 + rune(c-'a'+10)
			case c >= 'A' && c <= 'F':
				v = v*16 + rune(c-'A'+10)
			default:
				return v, k
			}
		}
		return v, k
	}
	for i < len(s) {
		c := s[i]
		if c == '\'' {
			return string(b), i + 1, true
		}
		if c != '\\' || i+1 >= len(s) {
			b = append(b, c)
			i++
			continue
		}
		n := s[i+1]
		i += 2
		switch n {
		case 'n':
			b = append(b, '\n')
		case 't':
			b = append(b, '\t')
		case 'r':
			b = append(b, '\r')
		case 'a':
			b = append(b, 7)
		case 'b':
			b = append(b, 8)
		case 'e', 'E':
			b = append(b, 27)
		case 'f':
			b = append(b, '\f')
		case 'v':
			b = append(b, '\v')
		case 'x':
			v, k := hex(2)
			if k == 0 {
				b = append(b, '\\', 'x')
			} else {
				b = append(b, byte(v))
				i += k
			}
		case 'u', 'U':
			max := 4
			if n == 'U' {
				max = 8
			}
			v, k := hex(max)
			if k == 0 {
				b = append(b, '\\', n)
			} else {
				b = utf8.AppendRune(b, v)
				i += k
			}
		case 'c':
			if i < len(s) {
				b = append(b, s[i]&0x1f)
				i++
			}
		case '0', '1', '2', '3', '4', '5', '6', '7':
			v := int(n - '0')
			for k := 0; k < 2 && i < len(s) && s[i] >= '0' && s[i] <= '7'; k++ {
				v = v*8 + int(s[i]-'0')
				i++
			}
			b = append(b, byte(v))
		case '\\', '\'', '"', '?':
			b = append(b, n)
		default:
			b = append(b, '\\', n)
		}
	}
	return "", len(s), false
}

func shNameStart(c byte) bool { return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func shNameChar(c byte) bool { return shNameStart(c) || c >= '0' && c <= '9' }

// shIsAssign reports a bash NAME=value (or NAME+=value, NAME[i]=value) word: its name and = unquoted.
func shIsAssign(b []byte, unq []bool) bool {
	if len(b) == 0 || !unq[0] || !shNameStart(b[0]) {
		return false
	}
	k := 1
	for k < len(b) && unq[k] && shNameChar(b[k]) {
		k++
	}
	if k < len(b) && unq[k] && b[k] == '[' {
		for k < len(b) && b[k] != ']' {
			k++
		}
		k++
	}
	if k < len(b) && unq[k] && b[k] == '+' {
		k++
	}
	return k < len(b) && unq[k] && b[k] == '='
}

// shBraces writes out bash's brace lists in a word ({a,b} and a{b,c}d; nested): the words it gives, empty ones
// dropped, or nil when the word has no list. ok is false when it would give more than shMaxBraceWords; seq is set
// when the word holds a sequence ({1..9}), which the check cannot name.
func shBraces(b []byte, unq []bool) (alts []string, ok, seq bool) {
	out, ok, seq, found := shBraceList(b, unq, 0)
	if !ok || seq || !found {
		return nil, ok, seq
	}
	alts = []string{}
	for _, w := range out {
		if w != "" {
			alts = append(alts, w)
		}
	}
	return alts, true, false
}

func shBraceList(b []byte, unq []bool, level int) (out []string, ok, seq, found bool) {
	if level > 16 {
		return nil, false, false, true
	}
	for p := 0; p < len(b); p++ {
		if b[p] != '{' || !unq[p] {
			continue
		}
		depth := 0
		var commas []int
		dots := false
		q := -1
		for k := p; k < len(b); k++ {
			if !unq[k] {
				continue
			}
			switch b[k] {
			case '{':
				depth++
			case '}':
				depth--
			case ',':
				if depth == 1 {
					commas = append(commas, k)
				}
			case '.':
				if depth == 1 && k+1 < len(b) && b[k+1] == '.' && unq[k+1] {
					dots = true
				}
			}
			if depth == 0 {
				q = k
				break
			}
		}
		if q < 0 {
			continue // an unmatched { is a plain character
		}
		if len(commas) == 0 {
			if dots {
				return nil, true, true, true
			}
			continue // {x} is not a list
		}
		edges := append(append([]int{p}, commas...), q)
		for e := 0; e+1 < len(edges); e++ {
			var nb []byte
			var nu []bool
			nb = append(append(append(nb, b[:p]...), b[edges[e]+1:edges[e+1]]...), b[q+1:]...)
			nu = append(append(append(nu, unq[:p]...), unq[edges[e]+1:edges[e+1]]...), unq[q+1:]...)
			sub, ok, s, f := shBraceList(nb, nu, level+1)
			if !ok {
				return nil, false, false, true
			}
			if s {
				return nil, true, true, true
			}
			if !f {
				sub = []string{string(nb)}
			}
			out = append(out, sub...)
			if len(out) > shMaxBraceWords {
				return nil, false, false, true
			}
		}
		return out, true, false, true
	}
	return nil, true, false, false
}

// psDash gives the length of a dash PowerShell reads as one at position i (-, en dash, em dash, horizontal bar),
// 0 for none.
func (sc *shScan) psDash(i int) int {
	if i >= len(sc.s) {
		return 0
	}
	if sc.s[i] == '-' {
		return 1
	}
	r, n := utf8.DecodeRuneInString(sc.s[i:])
	if r == '–' || r == '—' || r == '―' {
		return n
	}
	return 0
}

// psQuote reads the quote at position i: 1 for a single quote, 2 for a double one (the curly ones PowerShell
// reads as plain quotes included), 0 for none, and its length.
func (sc *shScan) psQuote(i int) (kind, n int) {
	if i >= len(sc.s) {
		return 0, 0
	}
	switch sc.s[i] {
	case '\'':
		return 1, 1
	case '"':
		return 2, 1
	}
	if sc.s[i] < utf8.RuneSelf {
		return 0, 0
	}
	r, size := utf8.DecodeRuneInString(sc.s[i:])
	switch r {
	case '‘', '’', '‚', '‛':
		return 1, size
	case '“', '”', '„':
		return 2, size
	}
	return 0, 0
}

// psWord reads one word by PowerShell's quoting.
func (sc *shScan) psWord(cmd *shCommand) shWord {
	const globs = "*?["
	start := sc.i
	var wb shBuild
	s := sc.s
	if n := sc.psDash(sc.i); n > 0 && sc.i+n < len(s) && (shNameStart(s[sc.i+n]) || s[sc.i+n] == '?') {
		wb.w.param = true
		wb.lit("-", false, "")
		sc.i += n
	}
	if sc.at("@'") || sc.at(`@"`) {
		if text, ok := sc.psHereString(cmd); ok {
			wb.lit(text, true, globs)
			wb.w.raw = s[start:sc.i]
			wb.w.text = string(wb.b)
			return wb.w
		}
		if sc.err != "" {
			return wb.w
		}
	}
loop:
	for sc.i < len(s) && sc.err == "" {
		if sc.blank(sc.i) > 0 {
			break
		}
		if kind, n := sc.psQuote(sc.i); kind == 1 {
			sc.i += n
			sc.psSingle(&wb)
			continue
		} else if kind == 2 {
			sc.i += n
			sc.psDouble(&wb, cmd)
			continue
		}
		c := s[sc.i]
		switch c {
		case '\n', ';', '|', ')', '}', '(', '{', '<', '>':
			break loop
		case '&':
			break loop
		case ',':
			wb.w.comma = true
			sc.i++
			break loop
		case '`':
			if sc.i+1 >= len(s) {
				sc.i++
				continue
			}
			_, n := utf8.DecodeRuneInString(s[sc.i+1:])
			wb.lit(s[sc.i+1:sc.i+1+n], true, globs)
			sc.i += 1 + n
		case '$':
			sc.psDollar(&wb, cmd)
		case '@':
			if len(wb.b) == 0 && sc.i+1 < len(s) && (s[sc.i+1] == '(' || s[sc.i+1] == '{') {
				p := sc.i
				closeBy := byte(')')
				if s[sc.i+1] == '{' {
					closeBy = '}'
				}
				sc.i += 2
				if l := sc.nested(closeBy); l != nil {
					cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
				}
				wb.exp(s[p:sc.i])
				continue
			}
			if len(wb.b) == 0 && sc.i+1 < len(s) && shNameStart(s[sc.i+1]) {
				j := sc.i + 1
				for j < len(s) && shNameChar(s[j]) {
					j++
				}
				wb.exp(s[sc.i:j])
				wb.w.splat = true
				sc.i = j
				continue
			}
			wb.lit("@", false, "")
			sc.i++
		default:
			_, n := utf8.DecodeRuneInString(s[sc.i:])
			wb.lit(s[sc.i:sc.i+n], false, globs)
			sc.i += n
		}
	}
	wb.w.raw = s[start:sc.i]
	wb.w.text = string(wb.b)
	return wb.w
}

// psSingle reads the inside of a single-quoted string: literal, a doubled quote standing for one.
func (sc *shScan) psSingle(wb *shBuild) {
	wb.w.quoted = true
	for sc.i < len(sc.s) {
		if kind, n := sc.psQuote(sc.i); kind == 1 {
			if k2, n2 := sc.psQuote(sc.i + n); k2 == 1 {
				wb.lit("'", true, "")
				sc.i += n + n2
				continue
			}
			sc.i += n
			return
		}
		_, n := utf8.DecodeRuneInString(sc.s[sc.i:])
		wb.lit(sc.s[sc.i:sc.i+n], true, "*?[")
		sc.i += n
	}
	sc.fail("a quote left open")
}

// psDouble reads the inside of a double-quoted string: the backtick escapes, $( ) is split where it stands, a
// variable marks the word.
func (sc *shScan) psDouble(wb *shBuild, cmd *shCommand) {
	if !sc.enter() {
		return
	}
	defer sc.leave()
	wb.w.quoted = true
	for sc.i < len(sc.s) && sc.err == "" {
		if kind, n := sc.psQuote(sc.i); kind == 2 {
			if k2, n2 := sc.psQuote(sc.i + n); k2 == 2 {
				wb.lit(`"`, true, "")
				sc.i += n + n2
				continue
			}
			sc.i += n
			return
		}
		switch c := sc.s[sc.i]; c {
		case '`':
			if sc.i+1 < len(sc.s) {
				_, n := utf8.DecodeRuneInString(sc.s[sc.i+1:])
				wb.lit(sc.s[sc.i+1:sc.i+1+n], true, "*?[")
				sc.i += 1 + n
			} else {
				sc.i++
			}
		case '$':
			sc.psDollar(wb, cmd)
		default:
			_, n := utf8.DecodeRuneInString(sc.s[sc.i:])
			wb.lit(sc.s[sc.i:sc.i+n], true, "*?[")
			sc.i += n
		}
	}
	sc.fail("a quote left open")
}

// psDollar reads what starts with $: $( ), ${...}, a variable ($x, $env:X, $_), or a plain $.
func (sc *shScan) psDollar(wb *shBuild, cmd *shCommand) {
	s, i := sc.s, sc.i
	if i+1 >= len(s) {
		wb.lit("$", false, "")
		sc.i++
		return
	}
	switch n := s[i+1]; {
	case n == '(':
		sc.i += 2
		if l := sc.nested(')'); l != nil {
			cmd.nested = append(cmd.nested, shNest{list: l, sub: true})
		}
		wb.exp(s[i:sc.i])
	case n == '{':
		end := strings.IndexByte(s[i+2:], '}')
		if end < 0 {
			sc.fail("a bracket left open")
			sc.i = len(s)
			return
		}
		sc.i = i + 2 + end + 1
		wb.exp(s[i:sc.i])
	case shNameChar(n) || n == '?' || n == '^' || n == '$' || n == ':':
		j := i + 1
		for j < len(s) && (shNameChar(s[j]) || s[j] == ':' || j == i+1 && (s[j] == '?' || s[j] == '^' || s[j] == '$')) {
			j++
		}
		wb.exp(s[i:j])
		sc.i = j
	default:
		wb.lit("$", false, "")
		sc.i++
	}
}

// psHereString reads a here-string, @'...'@ or @"..."@: its text. ok is false when the scanner does not stand on
// one (an @' with more on its line).
func (sc *shScan) psHereString(cmd *shCommand) (string, bool) {
	s := sc.s
	q := s[sc.i+1]
	j := sc.i + 2
	for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\r') {
		j++
	}
	if j >= len(s) || s[j] != '\n' {
		return "", false
	}
	body := j + 1
	closing := "\n" + string(q) + "@"
	var text string
	if strings.HasPrefix(s[body:], string(q)+"@") {
		text = ""
		sc.i = body + 2
	} else {
		end := strings.Index(s[body:], closing)
		if end < 0 {
			sc.fail("a here-string left open")
			sc.i = len(s)
			return "", false
		}
		text = s[body : body+end]
		sc.i = body + end + len(closing)
	}
	text = strings.TrimSuffix(text, "\r")
	if q == '"' {
		sub := &shScan{s: text, d: shPower, depth: sc.depth}
		var wb shBuild
		for sub.i < len(sub.s) && sub.err == "" {
			switch sub.s[sub.i] {
			case '`':
				sub.adv(2)
			case '$':
				sub.psDollar(&wb, cmd)
			default:
				sub.i++
			}
		}
		if sub.err != "" {
			sc.fail(sub.err)
		}
	}
	return text, true
}

// cmdWord reads one word by cmd.exe's quoting: double quotes, which a line end closes, and the caret.
func (sc *shScan) cmdWord(cmd *shCommand) shWord {
	const globs = "*?"
	start := sc.i
	var wb shBuild
	s := sc.s
	inQuote := false
loop:
	for sc.i < len(s) {
		c := s[sc.i]
		if !inQuote {
			if sc.blank(sc.i) > 0 {
				break
			}
			switch c {
			case '\n', '&', '|', '<', '>', '(', ')':
				break loop
			case '^':
				if sc.i+1 < len(s) {
					wb.lit(s[sc.i+1:sc.i+2], true, globs)
					sc.i += 2
				} else {
					sc.i++
				}
				continue
			}
		} else if c == '\n' {
			break
		}
		switch c {
		case '"':
			inQuote = !inQuote
			wb.w.quoted = true
			sc.i++
		case '%':
			sc.cmdPercent(&wb)
		case '!':
			if end := strings.IndexAny(s[sc.i+1:], "!\n"); end > 0 && s[sc.i+1+end] == '!' {
				wb.exp(s[sc.i : sc.i+end+2])
				sc.i += end + 2
				continue
			}
			wb.lit("!", inQuote, "")
			sc.i++
		default:
			wb.lit(s[sc.i:sc.i+1], inQuote, globs)
			sc.i++
		}
	}
	wb.w.raw = s[start:sc.i]
	wb.w.text = string(wb.b)
	return wb.w
}

// cmdPercent reads cmd.exe's % forms: %NAME%, %~dp0, %1, %*, a for loop's %%i or %i; a % alone stays.
func (sc *shScan) cmdPercent(wb *shBuild) {
	s, i := sc.s, sc.i
	if i+1 >= len(s) {
		wb.lit("%", false, "")
		sc.i++
		return
	}
	n := s[i+1]
	switch {
	case n == '%' && i+2 < len(s) && shNameChar(s[i+2]):
		wb.exp(s[i : i+3])
		sc.i = i + 3
		return
	case n == '~' || n == '*' || n >= '0' && n <= '9':
		j := i + 2
		for j < len(s) && shNameChar(s[j]) {
			j++
		}
		wb.exp(s[i:j])
		sc.i = j
		return
	}
	if end := strings.IndexAny(s[i+1:], "% \t\n&|<>()\""); end > 0 && s[i+1+end] == '%' {
		wb.exp(s[i : i+end+2])
		sc.i = i + end + 2
		return
	}
	if shNameStart(n) {
		wb.exp(s[i : i+2]) // a for loop's %i at the prompt
		sc.i = i + 2
		return
	}
	wb.lit("%", false, "")
	sc.i++
}

// shDecodeUTF16 reads PowerShell's -EncodedCommand text: base64 of UTF-16LE. ok is false when it is not.
func shDecodeUTF16(raw []byte) (string, bool) {
	if len(raw)%2 != 0 {
		return "", false
	}
	u := make([]uint16, len(raw)/2)
	for k := range u {
		u[k] = uint16(raw[2*k]) | uint16(raw[2*k+1])<<8
	}
	return string(utf16.Decode(u)), true
}
