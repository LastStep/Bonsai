package reader

// Cases beyond the formats set: the edges of each rule, this reader's own two codes, dispatch, line numbers and
// messages. Each row is a file (YAML unless md is set) and its outcome: a value as JSON, a code, or format-0.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

type tcase struct {
	name string
	in   string
	md   bool
	want string // "format-0", a reason code, or the accepted value as one line of JSON
	line int    // for a refusal, the line it names (0: not checked)
}

const f1 = "format: bonsai.task/1\n"

// key1024 is a key of 1024 characters, the longest both libraries read.
var key1024 = "k" + strings.Repeat("x", 1023)

var cases = []tcase{
	// Dispatch.
	{"empty file", "", false, "format-0", 0},
	{"only comments", "# a\n\n# b\n", false, "format-0", 0},
	{"md without frontmatter", "# Title\n\nformat: bonsai.task/1\n", true, "format-0", 0},
	{"md never closed", "---\nformat: bonsai.task/1\nid: x\n", true, "format-0", 0},
	{"md empty frontmatter", "---\n---\nbody\n", true, "format-0", 0},
	{"md opener with a trailing space", "--- \nformat: bonsai.task/1\n---\n", true, "format-0", 0},
	{"md opener alone at the end", "---", true, "format-0", 0},
	{"md closer at the end of the file", "---\nformat: bonsai.task/1\n---", true, `{"format":"bonsai.task/1"}`, 0},
	{"md body not read", "---\nformat: bonsai.task/1\n---\n\x00 \t bad: [ \"\n", true, `{"format":"bonsai.task/1"}`, 0},
	{"md --- with a space inside", "---\nformat: bonsai.task/1\n--- \n---\n", true, "doc-marker", 3},
	{"quoted format key first", "\"format\": bonsai.task/1\n", false, "key-quoted", 1},
	{"format not first wins over everything", "Bad Key: 1\n\tx\nformat: bonsai.task/1\n", false, "format-not-first", 3},
	{"format: with a space is no format key", "format : bonsai.task/1\nid: x\n", false, "format-0", 0},
	{"indented top level", "  format: bonsai.task/1\n  id: T-1\n", false, `{"format":"bonsai.task/1","id":"T-1"}`, 0},
	{"indented top level, a line less indented", "  format: bonsai.task/1\nid: T-1\n", false, "line-not-read", 2},
	{"format too new", "format: bonsai.task/2\nid: [\n", false, "format-too-new", 1},
	{"format too new, quoted", "format: \"bonsai.task/12\"\n", false, "format-too-new", 1},
	{"format too new before a doc marker", "---\nformat: bonsai.task/2\n", false, "format-too-new", 2},
	{"format with no major is read", "format: post\n", false, `{"format":"post"}`, 0},
	{"format twice", f1 + "format: bonsai.task/1\n", false, "key-twice", 2},
	{"BOM then comment", "\xEF\xBB\xBF# c\n" + f1, false, `{"format":"bonsai.task/1"}`, 0},
	{"two BOMs", "\xEF\xBB\xBF\xEF\xBB\xBF" + f1, false, "format-0", 0},

	// Lines: text, tabs, document markers, every line read.
	{"invalid UTF-8", f1 + "id: \xff\n", false, "not-text", 2},
	{"a NUL", f1 + "id: a\x00b\n", false, "not-text", 2},
	{"a lone CR", f1 + "id: a\rb\n", false, "not-text", 2},
	{"a CR at the end with no LF", f1 + "id: a\r", false, "not-text", 2},
	{"a DEL", f1 + "id: a\x7f\n", false, "not-text", 2},
	{"a C1 control", f1 + "id: a\xc2\x90\n", false, "not-text", 2},
	{"U+2028 in a comment", f1 + "# a\xe2\x80\xa8b\n", false, "not-text", 2},
	{"NEL in quotes", f1 + "id: \"a\xc2\x85b\"\n", false, "not-text", 2},
	{"U+FFFE", f1 + "id: a\xef\xbf\xbe\n", false, "not-text", 2},
	{"a BOM inside text", f1 + "id: a\xef\xbb\xbfb\n", false, `{"format":"bonsai.task/1","id":"a\ufeffb"}`, 0},
	{"non-ASCII text", f1 + "id: Caf\xc3\xa9\xc2\xa0x\n", false, `{"format":"bonsai.task/1","id":"Caf\u00e9\u00a0x"}`, 0},
	{"not-text before a later problem", f1 + "a: \x01\nB: 1\n", false, "not-text", 2},
	{"key problem before a later not-text", f1 + "B: 1\na: \x01\n", false, "key-form", 2},
	{"a tab-only blank line", f1 + "\t\nid: x\n", false, "tab-indent", 2},
	{"a tab after spaces", f1 + "a:\n  \tb: 1\n", false, "tab-indent", 3},
	{"a tab-indented comment", f1 + "\t# c\n", false, "tab-indent", 2},
	{"a tab in a plain value", f1 + "id: a\tb\n", false, "quote-this-value", 2},
	{"a tab after a plain value", f1 + "id: a\t\n", false, "quote-this-value", 2},
	{"a tab after the colon", f1 + "id:\tvalue\n", false, "quote-this-value", 2},
	{"a tab after the colon, nothing after", f1 + "id:\t\n", false, "quote-this-value", 2},
	{"a tab after an item key's colon", f1 + "l:\n  - k:\tv\n", false, "quote-this-value", 3},
	{"format: and a tab is the format key", "format:\tbonsai.task/1\nid: x\n", false, "quote-this-value", 1},
	{"format: and a tab, too new", "format:\tbonsai.task/2\n", false, "format-too-new", 1},
	{"format: and a tab, not first", "id: x\nformat:\tbonsai.task/1\n", false, "format-not-first", 2},
	{"a tab in double quotes", f1 + "id: \"a\tb\"\n", false, `{"format":"bonsai.task/1","id":"a\tb"}`, 0},
	{"a tab in a comment", f1 + "id: a # c\td\n", false, `{"format":"bonsai.task/1","id":"a"}`, 0},
	{"a tab inside block text", f1 + "t: |\n  a\tb\n", false, `{"format":"bonsai.task/1","t":"a\tb\n"}`, 0},
	{"a tab after a closing quote", f1 + "id: \"a\"\t# c\n", false, "after-quote", 2},
	{"--- with text after it", f1 + "--- x\n", false, "doc-marker", 2},
	{"an indented --- is no marker", f1 + "  ---\n", false, "line-not-read", 2},
	{"a value on the next line", f1 + "id:\n  value\n", false, "line-not-read", 3},
	{"a plain value over two lines", f1 + "id: a\n  b\n", false, "line-not-read", 3},
	{"a deeper line after a scalar", f1 + "a: 1\n    b: 2\n", false, "line-not-read", 3},
	{"a line between two indentations", f1 + "a:\n    b: 1\n  c: 2\n", false, "line-not-read", 4},
	{"a key: with no space", f1 + "id:x\n", false, "line-not-read", 2},
	{"a stray value", f1 + "just text\n", false, "line-not-read", 2},
	{"a quoted line alone", f1 + "\"just text\"\n", false, "line-not-read", 2},

	// Keys.
	{"a reserved word in another case", f1 + "Yes: 1\n", false, "key-form", 2},
	{"an anchor on a key", f1 + "&a k: v\n", false, "key-form", 2},
	{"a space before the colon", f1 + "id : x\n", false, "key-form", 2},
	{"two dots", f1 + "a.b.c: 1\n", false, "key-form", 2},
	{"a label key", f1 + "ns-1.k_2: 1\n", false, `{"format":"bonsai.task/1","ns-1.k_2":1}`, 0},
	{"a label key with an upper case letter", f1 + "ns.K: 1\n", false, "key-form", 2},
	{"<< with a space", f1 + "<< : x\n", false, "key-form", 2},
	{"? alone", f1 + "?\n", false, "key-complex", 2},
	{"?x is a key form", f1 + "?x: 1\n", false, "key-form", 2},
	{"a quoted key beats its bad escape", f1 + "\"a\\q\": 1\n", false, "key-quoted", 2},
	{"a key twice inside an item", f1 + "l:\n  - a: 1\n    a: 2\n", false, "key-twice", 4},
	{"a key: in a comment", f1 + "a # b: c\n", false, "line-not-read", 2},

	// Key length: §2.4 fixes no length; both libraries refuse a key over 1024 characters.
	{"a key of 1024 characters", f1 + key1024 + ": 1\n", false, `{"format":"bonsai.task/1","` + key1024 + `":1}`, 0},
	{"a key of 1025 characters", f1 + key1024 + "x: 1\n", false, "key-form", 2},
	{"a nested key of 1025 characters", f1 + "a:\n  " + key1024 + "x: 1\n", false, "key-form", 3},
	{"an item key of 1025 characters", f1 + "a:\n  - " + key1024 + "x: 1\n", false, "key-form", 3},
	{"a label key of 1025 characters", f1 + "a:\n  ns." + key1024[:1022] + ": 1\n", false, "key-form", 3},
	{"a label key of 1024 characters", f1 + "a:\n  ns." + key1024[:1021] + ": 1\n", false, `{"format":"bonsai.task/1","a":{"ns.` + key1024[:1021] + `":1}}`, 0},

	// A ? inside a flow item: a YAML 1.1 library ends a plain scalar there.
	{"a ? ending a flow item", f1 + "l: [what?]\n", false, "quote-this-value", 2},
	{"a ? inside a flow item", f1 + "l: [x/?q, b]\n", false, "quote-this-value", 2},
	{"a ? in block text", f1 + "a: what? x?y\n", false, `{"format":"bonsai.task/1","a":"what? x?y"}`, 0},
	{"a quoted ? in a flow", f1 + "l: [\"what?\"]\n", false, `{"format":"bonsai.task/1","l":["what?"]}`, 0},

	// Order on one line: each problem where it starts; not-text at its character; of two at one character, the
	// code listed first.
	{"the key before a bad character", f1 + "B: \x01\n", false, "key-form", 2},
	{"a bad character inside a key", f1 + "a\x01: 1\n", false, "key-form", 2},
	{"a bad character starting a key", f1 + "\x01a: 1\n", false, "not-text", 2},
	{"a scalar that fits no row before a bad character", f1 + "k: 0\x01\n", false, "quote-this-value", 2},
	{"a bad character in text", f1 + "k: a\x01b\n", false, "not-text", 2},
	{"a bad character before a bad escape", f1 + "k: \"a\x01\\q\"\n", false, "not-text", 2},
	{"a bad escape before a bad character", f1 + "k: \"\\qa\x01\"\n", false, "bad-escape", 2},
	{"an unclosed quote at its opening", f1 + "k: \"a\\q\n", false, "quoted-multiline", 2},
	{"an unclosed quote before a bad character", f1 + "k: \"a\x01\n", false, "quoted-multiline", 2},
	{"flow items before what follows the ]", f1 + "l: [&x] y\n", false, "anchor", 2},
	{"flow items left to right", f1 + "l: [a\x01, 0755]\n", false, "not-text", 2},
	{"flow items left to right, the first first", f1 + "l: [0755, a\x01]\n", false, "quote-this-value", 2},
	{"text after the ] before a bad character", f1 + "l: [a] y\x01\n", false, "line-not-read", 2},
	{"an unclosed flow at its [", f1 + "l: [a, 0755\n", false, "flow-multiline", 2},
	{"an unclosed flow before an anchor", f1 + "l: [&a, b\n", false, "flow-multiline", 2},
	{"a bad header before a bad character", f1 + "t: |\x01\n", false, "block-indicator", 2},
	{"a bad character in a header's comment", f1 + "t: | # c\x01\n  a\n", false, "not-text", 2},
	{"a bad character in a comment before nested lines", f1 + "a: # c\x01\n  - 0755\n", false, "not-text", 2},
	{"a # line before its bad character", f1 + "t: |\n  #\x01\n", false, "block-hash-line", 3},
	{"a bad character in block text", f1 + "t: |\n  a\x01\n", false, "not-text", 3},
	{"a dash and two spaces before a bad character", f1 + "l:\n  -  a\x01\n", false, "seq-dash-space", 3},
	{"an anchor before a bad character", f1 + "k: &a\x01\n", false, "anchor", 2},
	{"{} then text before a bad character", f1 + "m: {} x\x01\n", false, "line-not-read", 2},

	// Block sequences.
	{"a - alone", f1 + "l:\n  -\n", false, "seq-dash-space", 3},
	{"a - and a tab", f1 + "l:\n  -\tx\n", false, "seq-dash-space", 3},
	{"one space and a tab", f1 + "l:\n  - \tx\n", false, "seq-dash-space", 3},
	{"an empty item", f1 + "l:\n  - \n  - a\n", false, `{"format":"bonsai.task/1","l":[null,"a"]}`, 0},
	{"an empty item over a mapping", f1 + "l:\n  - \n    k: v\n", false, `{"format":"bonsai.task/1","l":[{"k":"v"}]}`, 0},
	{"an item with a comment over a sequence", f1 + "l:\n  - # c\n    - x\n", false, `{"format":"bonsai.task/1","l":[["x"]]}`, 0},
	{"- - a", f1 + "l:\n  - - a\n", false, "quote-this-value", 3},
	{"a mapping item", f1 + "l:\n  - a: 1\n    b: [x]\n    c:\n      d: e\n  - z\n", false,
		`{"format":"bonsai.task/1","l":[{"a":1,"b":["x"],"c":{"d":"e"}},"z"]}`, 0},
	{"an item key at the wrong place", f1 + "l:\n  - a: 1\n     b: 2\n", false, "line-not-read", 4},
	{"a mapping where an item belongs", f1 + "l:\n  - a\n  b: 1\n", false, "line-not-read", 4},
	{"a deeper item", f1 + "l:\n  - a\n    - b\n", false, "line-not-read", 4},
	{"a sequence at its key's own indentation in an item", f1 + "l:\n  - k:\n    - a\n", false, "line-not-read", 4},
	{"an item holding : ", f1 + "l:\n  - Rule: no tabs\n", false, "key-form", 3},
	{"an item quoted with a colon inside", f1 + "l:\n  - \"a: b\"\n", false, `{"format":"bonsai.task/1","l":["a: b"]}`, 0},
	{"a quoted key in an item", f1 + "l:\n  - \"a\": b\n", false, "key-quoted", 3},
	{"a flow mapping item", f1 + "l:\n  - {a: 1}\n", false, "flow-mapping", 3},
	{"a block scalar item", f1 + "l:\n  - |\n   x\n  - z\n", false, `{"format":"bonsai.task/1","l":["x\n","z"]}`, 0},
	{"a sequence of sequences' items at depth", f1 + "a:\n  b:\n    - 1\n    - 2\n  c: 3\n", false,
		`{"format":"bonsai.task/1","a":{"b":[1,2],"c":3}}`, 0},

	// Flow sequences and {}.
	{"a trailing comma", f1 + "l: [a, ]\n", false, "line-not-read", 2},
	{"two commas", f1 + "l: [a,,b]\n", false, "line-not-read", 2},
	{"text after the ]", f1 + "l: [a] x\n", false, "line-not-read", 2},
	{"a comment after the ]", f1 + "l: [a]  # c ]\n", false, `{"format":"bonsai.task/1","l":["a"]}`, 0},
	{"a ] in a comment", f1 + "l: [a #c]\n", false, "flow-multiline", 2},
	{"a quoted item never closed", f1 + "l: [\"a, b]\n", false, "flow-multiline", 2},
	{"a flow mapping item in a flow", f1 + "l: [a, {b: 1}]\n", false, "flow-mapping", 2},
	{"{} in a flow", f1 + "l: [a, {}]\n", false, "flow-nested", 2},
	{"a [ inside an item", f1 + "l: [a[b]\n", false, "flow-nested", 2},
	{"a } then a [ in an item", f1 + "l: [a}b[c]\n", false, "quote-this-value", 2},
	{"junk after a quoted item", f1 + "l: [\"a\" b, c]\n", false, "after-quote", 2},
	{"a quote inside a quoted item", f1 + "l: [\"a\"b\", c]\n", false, "unescaped-quote", 2},
	{"an anchor in a flow", f1 + "l: [&a x]\n", false, "anchor", 2},
	{"an alias in a flow", f1 + "l: [*a]\n", false, "alias", 2},
	{"a tag in a flow", f1 + "l: [!t x]\n", false, "tag", 2},
	{"a colon inside an item", f1 + "l: [a:b, 'c: d']\n", false, `{"format":"bonsai.task/1","l":["a:b","c: d"]}`, 0},
	{"an item holding : ", f1 + "l: [a: b]\n", false, "quote-this-value", 2},
	{"spaces inside", f1 + "l: [ ]\nm: { }\n", false, `{"format":"bonsai.task/1","l":[],"m":{}}`, 0},
	{"[] then text", f1 + "l: [] x\n", false, "line-not-read", 2},
	{"{} then text", f1 + "m: {} x\n", false, "line-not-read", 2},
	{"{ never closed", f1 + "m: {\n", false, "flow-mapping", 2},
	{"typed items", f1 + "l: [1, -2.50, ~, true, null, '', \"\", 2026-10-08, x y]\n", false,
		`{"format":"bonsai.task/1","l":[1,-2.50,null,true,null,"","","2026-10-08","x y"]}`, 0},
	{"an escaped quote in a flow item", f1 + "l: [\"a\\\"b\", 'c''d']\n", false, `{"format":"bonsai.task/1","l":["a\"b","c'd"]}`, 0},

	// Block scalars.
	{"a comment after the header", f1 + "t: | # c\n  a\n", false, `{"format":"bonsai.task/1","t":"a\n"}`, 0},
	{"# right after the header", f1 + "t: |#\n  a\n", false, "block-indicator", 2},
	{"|2-", f1 + "t: |2-\n  a\n", false, "block-indicator", 2},
	{"text after the header", f1 + "t: | x\n", false, "block-indicator", 2},
	{"an empty block", f1 + "t: |\nm: 1\n", false, `{"format":"bonsai.task/1","t":"","m":1}`, 0},
	{"an empty block over blank lines", f1 + "t: >\n\n\nm: 1\n", false, `{"format":"bonsai.task/1","t":"","m":1}`, 0},
	{"leading empty lines", f1 + "t: |\n\n  a\n", false, `{"format":"bonsai.task/1","t":"\na\n"}`, 0},
	{"folded leading empty lines", f1 + "t: >\n\n  a\n  b\n\n\n  c\n", false, `{"format":"bonsai.task/1","t":"\na b\n\nc\n"}`, 0},
	{"folded trailing spaces kept", f1 + "t: >\n  a  \n  b\n", false, `{"format":"bonsai.task/1","t":"a   b\n"}`, 0},
	{"a blank line of spaces deeper than the block", f1 + "t: |\n  a\n     \n", false, "line-not-read", 4},
	{"a leading blank line deeper than the first line", f1 + "t: |\n     \n  a\n", false, "line-not-read", 3},
	{"a blank line of spaces within the block", f1 + "t: |\n  a\n  \n  b\n", false, `{"format":"bonsai.task/1","t":"a\n\nb\n"}`, 0},
	{"a comment at the margin ends the block", f1 + "t: |\n  a\n# c\nm: 1\n", false, `{"format":"bonsai.task/1","t":"a\n","m":1}`, 0},
	{"a deeper # line", f1 + "t: |\n  a\n    # b\n", false, "block-hash-line", 4},
	{"a # first line", f1 + "t: |\n  # a\n", false, "block-hash-line", 3},
	{"a > line deeper after an empty one", f1 + "t: >-\n  a\n\n   b\n", false, "folded-deeper", 5},
	{"a literal keeps deeper lines", f1 + "t: |-\n  a\n    b\n", false, `{"format":"bonsai.task/1","t":"a\n  b"}`, 0},
	{"clip at the end of the file", f1 + "t: |\n  a", false, "not-text", 3},
	{"folded clip at the end of the file", f1 + "t: >\n  a", false, "not-text", 3},
	{"strip at the end of the file", f1 + "t: |-\n  a", false, `{"format":"bonsai.task/1","t":"a"}`, 0},
	{"a block in an item mapping needs deeper lines", f1 + "l:\n  - k: |\n    x\n", false, "line-not-read", 4},
	{"a block in an item mapping", f1 + "l:\n  - k: |\n     x\n", false, `{"format":"bonsai.task/1","l":[{"k":"x\n"}]}`, 0},
	{"a less indented line ends the block", f1 + "a:\n  t: |\n      x\n    y\n", false, "line-not-read", 5},
	{"a block's --- line", f1 + "t: |\n  a\n---\n", false, "doc-marker", 4},
	{"a CRLF block", "format: bonsai.task/1\r\nt: >\r\n  a\r\n  b\r\n", false, `{"format":"bonsai.task/1","t":"a b\n"}`, 0},

	// Quoted scalars.
	{"a backslash at the end", f1 + "t: \"a\\\n", false, "quoted-multiline", 2},
	{"'' then nothing", f1 + "t: 'a''\n", false, "quoted-multiline", 2},
	{"a comment right after the quote", f1 + "t: \"x\"#c\n", false, "after-quote", 2},
	{"a quote inside the comment", f1 + "t: \"x\" # \"y\"\n", false, `{"format":"bonsai.task/1","t":"x"}`, 0},
	{"two single-quoted", f1 + "t: 'a' 'b'\n", false, "after-quote", 2},
	{"a backslash in single quotes", f1 + "t: 'a\\x'\n", false, `{"format":"bonsai.task/1","t":"a\\x"}`, 0},
	{"empty quotes", f1 + "a: ''\nb: \"\"\n", false, `{"format":"bonsai.task/1","a":"","b":""}`, 0},
	{"a bad escape before an unescaped quote", f1 + "t: \"a\\q\" b\"\n", false, "bad-escape", 2},

	// Plain scalars.
	{"-0 and decimals", f1 + "a: -0\nb: -0.0\nc: 10.500\n", false, `{"format":"bonsai.task/1","a":0,"b":-0.0,"c":10.500}`, 0},
	{"a time with seconds", f1 + "a: 2026-10-08 14:08:00\n", false, "quote-this-value", 2},
	{"a date with one-digit parts", f1 + "a: 2026-1-8\n", false, "quote-this-value", 2},
	{"Null and NULL", f1 + "a: Null\n", false, "quote-this-value", 2},
	{"FALSE", f1 + "a: FALSE\n", false, "quote-this-value", 2},
	{"OFF", f1 + "a: OFF\n", false, "quote-this-value", 2},
	{"y", f1 + "a: y\n", false, "quote-this-value", 2},
	{"a sexagesimal", f1 + "a: 1:20\n", false, "quote-this-value", 2},
	{"+1", f1 + "a: +1\n", false, "quote-this-value", 2},
	{"1.", f1 + "a: 1.\n", false, "quote-this-value", 2},
	{".5", f1 + "a: .5\n", false, "quote-this-value", 2},
	{"- as a value", f1 + "a: -\n", false, "quote-this-value", 2},
	{"- a as a value", f1 + "a: - a\n", false, "quote-this-value", 2},
	{"a non-ASCII first letter", f1 + "a: \xc3\xa9t\xc3\xa9\n", false, "quote-this-value", 2},
	{"a value ending in : after spaces", f1 + "a: x:   \n", false, "quote-this-value", 2},
	{"a > at the start is a header", f1 + "a: >5 apples\n", false, "block-indicator", 2},
	{"text with brackets outside a flow", f1 + "a: x [y] {z}, w\n", false, `{"format":"bonsai.task/1","a":"x [y] {z}, w"}`, 0},
	{"text with : inside", f1 + "a: http://x/y\n", false, `{"format":"bonsai.task/1","a":"http://x/y"}`, 0},
	{"text and many spaces", f1 + "a: x   y   # c\n", false, `{"format":"bonsai.task/1","a":"x   y"}`, 0},
	{"an anchor value", f1 + "a: &x\n", false, "anchor", 2},
	{"15 digits negative", f1 + "a: -999999999999999\n", false, `{"format":"bonsai.task/1","a":-999999999999999}`, 0},
	{"16 digits negative", f1 + "a: -1000000000000000\n", false, "quote-this-value", 2},
}

func TestReaderCases(t *testing.T) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var r Result
			if c.md {
				r = ReadMarkdown([]byte(c.in))
			} else {
				r = ReadYAML([]byte(c.in))
			}
			switch r.Outcome {
			case Format0:
				if c.want != "format-0" {
					t.Fatalf("format-0, want %s", c.want)
				}
			case Refused:
				if r.Refusal.Code != c.want {
					t.Fatalf("refused %v, want %s", r.Refusal, c.want)
				}
				if c.line != 0 && r.Refusal.Line != c.line {
					t.Errorf("line %d, want %d (%v)", r.Refusal.Line, c.line, r.Refusal)
				}
			case Accepted:
				want, err := schema.Decode([]byte(c.want))
				if err != nil {
					t.Fatalf("accepted %s, want %s", schema.Show(JSON(r.Value)), c.want)
				}
				got := JSON(r.Value)
				if !schema.Equal(got, want) || !sameOrder(got, want) {
					t.Fatalf("value %s, want %s", schema.Show(got), c.want)
				}
			}
		})
	}
}

// A refusal reads as one ASCII line naming its line, its code and a next step, whatever the file held.
func TestRefusalMessages(t *testing.T) {
	r := ReadYAML([]byte(f1 + "title: caf\xc3\xa9: x \x22\n"))
	if r.Outcome != Refused {
		t.Fatalf("outcome %s", r.Outcome)
	}
	msg := r.Refusal.Error()
	for i := 0; i < len(msg); i++ {
		if msg[i] < 0x20 || msg[i] > 0x7e {
			t.Fatalf("byte %d of %q is not printable ASCII", i, msg)
		}
	}
	if !strings.HasPrefix(msg, "line 2: ") || !strings.Contains(msg, "(quote-this-value); next: quote the value") {
		t.Errorf("message %q", msg)
	}
	if r.Err() == nil || (Result{Outcome: Accepted}).Err() != nil {
		t.Errorf("Err does not follow the outcome")
	}
	long := ReadYAML([]byte(f1 + "t: " + strings.Repeat("\xc3\xa9", 80) + "\n"))
	if long.Refusal == nil || !strings.Contains(long.Refusal.Message, `\u00e9..."`) || len(long.Refusal.Message) > 250 {
		t.Errorf("a long value is not cut in the message: %v", long.Refusal)
	}
}

// Formats set 4: a refusal names the next step that fits it. A file whose lines end in a CR alone is one line to
// format 1, so its refusal names the line endings, whatever its code (trick/yaml-1/lines-cr), and only once; an empty
// flow sequence item, refused line-not-read, names its own next step, not the indentation one (the question 5.1.2's
// builder raised); a line with no lone CR keeps its code's next step.
func TestNextStepsThatFit(t *testing.T) {
	cases := []struct {
		name, in, code, next, message string
	}{
		{"lone CR", "format: bonsai.lanes/1\rlanes: []\r", CodeQuoteThisValue, nextCREnds, "a CR that ends no line"},
		{"lone CR, two keys", "format: bonsai.lanes/1\rmode: 0755\r", CodeQuoteThisValue, nextCREnds,
			"a CR alone"},
		{"lone CR after a quoted value", "format: bonsai.lanes/1\nnote: \"a\" b\rc: d\n", CodeAfterQuote, nextCREnds,
			"a CR that ends no line"},
		{"trailing comma", "format: bonsai.lanes/1\nitems: [a, ]\n", CodeLineNotRead, nextEmptyFlowItem, "an empty item"},
		{"empty item mid-list", "format: bonsai.lanes/1\nitems: [a, , b]\n", CodeLineNotRead, nextEmptyFlowItem, "an empty item"},
		{"a plain refusal", "format: bonsai.lanes/1\nmode: 0755\n", CodeQuoteThisValue, next[CodeQuoteThisValue], "does not start"},
		{"a line not read", "format: bonsai.lanes/1\nitems:\n- a\n", CodeLineNotRead, next[CodeLineNotRead], "where a key belongs"},
	}
	for _, c := range cases {
		r := ReadYAML([]byte(c.in))
		if r.Outcome != Refused {
			t.Errorf("%s: outcome %s, want refused", c.name, r.Outcome)
			continue
		}
		if r.Refusal.Code != c.code || r.Refusal.Next != c.next || !strings.Contains(r.Refusal.Message, c.message) {
			t.Errorf("%s: %v\nwant code %s, next %q, a message holding %q", c.name, r.Refusal, c.code, c.next, c.message)
		}
		if strings.Count(r.Refusal.Message, "ends no line") > 1 {
			t.Errorf("%s: the line endings are named twice: %v", c.name, r.Refusal)
		}
	}
}

// Formats set 4: a quoted "format": and a tab is the file's first key, as format: and a tab is and as format 0's
// reader reads it; a quoted key followed by anything else after its colon is no key to dispatch.
func TestDispatchQuotedKeyAndTab(t *testing.T) {
	cases := []struct {
		in   string
		want Outcome
		code string
	}{
		{"\"format\":\tbonsai.lanes/1\nlanes: []\n", Refused, CodeKeyQuoted},
		{"'format':\tbonsai.lanes/1\nlanes: []\n", Refused, CodeKeyQuoted},
		{"\"format\": bonsai.lanes/1\nlanes: []\n", Refused, CodeKeyQuoted},
		{"\"format\":\n", Refused, CodeKeyQuoted},
		{"\"format\":x\nlanes: []\n", Format0, ""},
		{"format:\tbonsai.lanes/1\nlanes: []\n", Refused, CodeQuoteThisValue},
	}
	for _, c := range cases {
		r := ReadYAML([]byte(c.in))
		if r.Outcome != c.want || (c.code != "" && r.Refusal.Code != c.code) {
			t.Errorf("%q: %s %v, want %s %s", c.in, r.Outcome, r.Err(), c.want, c.code)
		}
	}
}

// Map's accessors, and JSON of every value kind.
func TestMapAndJSON(t *testing.T) {
	r := ReadYAML([]byte(f1 + "a: 1\nb:\n  c: [x, 2.5]\nd: |\n  t\n"))
	if r.Outcome != Accepted {
		t.Fatal(r.Err())
	}
	m := r.Value
	if m.Len() != 4 || len(m.Entries()) != 4 {
		t.Errorf("Len %d", m.Len())
	}
	if e, ok := m.Entry("b"); !ok || e.Line != 3 {
		t.Errorf("Entry(b) = %v, %v", e, ok)
	}
	if v, ok := m.Get("a"); !ok || v != int64(1) {
		t.Errorf("Get(a) = %#v", v)
	}
	if _, ok := m.Get("zz"); ok {
		t.Errorf("Get of a missing key")
	}
	b, _ := m.Get("b")
	c, _ := b.(*Map).Get("c")
	if d := c.([]any)[1].(Decimal); d.Float64() != 2.5 {
		t.Errorf("Decimal %v", d)
	}
	var nilMap *Map
	if nilMap.Len() != 0 || nilMap.Entries() != nil {
		t.Errorf("a nil Map is not empty")
	}
	if _, ok := nilMap.Get("a"); ok {
		t.Errorf("a nil Map holds a key")
	}
	// A float64 is a format-0 number; any other type is a bug in the caller.
	if JSON(1.5) != json.Number("1.5") {
		t.Errorf("JSON of a float64: %#v", JSON(1.5))
	}
	defer func() {
		if recover() == nil {
			t.Errorf("JSON of a float32 did not panic")
		}
	}()
	JSON(float32(1.5))
}

// The outcome names are expect.json's.
func TestOutcomeNames(t *testing.T) {
	for o, want := range map[Outcome]string{Accepted: "accepted", Refused: "refused", Format0: "format-0", 9: "unknown"} {
		if o.String() != want {
			t.Errorf("%d: %s, want %s", o, o, want)
		}
	}
}
