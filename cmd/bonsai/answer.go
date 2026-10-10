package main

// bonsai answer (spec section 4, section 8; contract section 9.3; plan-5 5.2.5): a person answers an open ask at a
// terminal, or a program records an answer a person gave elsewhere (--by its own reference, --via who carried it:
// Bonsai stores what it is given, and its code names no caller). A thin command over internal/asks; with --json it
// prints bonsai.asks/1 (format.Asks). Its table is answerWord. An answer answers; it never grants.

import (
	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/redact"
)

func init() { register(answerWord) }

var answerWord = &Word{
	Name:    "answer",
	Order:   9,
	Title:   "answer an open ask (contract section 9.3); an answer never grants anything.",
	Summary: "answer an open ask",
	Args:    "<key> [flags]",
	About: `It appends one answer record to the main checkout's .bonsai/local/asks/ and one ask record to the log. A
Decide is answered with --choice (one of its options, as written), a Look with --verdict, any ask with --words. The
first answer stands: the same answer again (the same key and --by) writes nothing and exits 0; any other answer to
an ask that is not open is refused. The session that filed the ask cannot answer it (CLAUDE_CODE_SESSION_ID). Every
text is checked as ask checks it: no hidden or control character, every secret redacted, its limit never cut.
`,
	Flags: []Flag{
		{Name: "--choice", Value: "TEXT", Help: "a Decide's option, as written (bonsai ask --status lists them)"},
		{Name: "--verdict", Value: "pass|fail", Help: "a Look's verdict"},
		{Name: "--words", Value: "TEXT", Help: "the person's own words, at most 2,000 characters (they may hold line feeds)"},
		{Name: "--by", Value: "TEXT", Help: "who answered: terminal unless given; a caller's own reference, at most 60 characters"},
		{Name: "--via", Value: "TEXT", Help: "who carried the answer, at most 30 characters; null unless given"},
		{Name: "--json", Help: "print the bonsai.asks/1 document (for programs) instead of plain text"},
	},
	Exits: []Exit{
		{Code: 0, Means: "answered, or the same answer was there already (nothing written)"},
		{Code: 2, Means: "bad input: no key, a flag answer does not take, no --choice, --verdict or --words, a choice not among\n" +
			"the options, a value broken (the message names the rule); bonsai.yaml refused. Nothing was written"},
		{Code: 3, Means: "the answer could not be written or the asks read (partly-written: the answer record stands, its log\nrecord not)"},
		{Code: 4, Means: "not in a git checkout or not linked (bonsai init); the key has no ask, or its ask is answered or\n" +
			"resolved (ask-not-open); the session that asked is answering (answer-own-session). Nothing was written"},
	},
	Examples: []string{
		`bonsai answer agent:h-3e9f0a1b2c4d --choice "Darker blue" --words "Darker reads better on old monitors."`,
		"bonsai answer agent:colour-choice --by act:4711 --via desk --verdict pass --json",
	},
	Refused: asksRefused,
	Run:     runAnswer,
}

func runAnswer(c *call) int {
	switch {
	case len(c.rest) == 0:
		return c.refuse(&engine.Error{Code: "missing-value", Exit: exitInput, What: "answer needs the ask's key",
			Next: "run `bonsai asks` to list the open asks, then `bonsai answer <key> ...`"})
	case len(c.rest) > 1:
		return c.refuse(c.flagError("answer takes one key, and no %+q", redact.Text(c.rest[1])))
	}
	site, e := askSite()
	if e != nil {
		return c.fail(e)
	}
	done, ae := asks.Answer(site.place, c.rest[0], asks.Answering{Choice: c.value("--choice"), Verdict: c.value("--verdict"),
		Words: c.value("--words"), By: c.value("--by"), Via: c.value("--via")})
	if ae != nil {
		return c.fail(askError(ae))
	}
	return c.asksOut(site, done, doneText(done, "answered"))
}
