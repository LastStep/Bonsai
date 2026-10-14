package guard

import "testing"

// The splitter keeps the redirections that write, with their targets (the roles' write check, rolecmds.go), and
// whether a here-document's delimiter was quoted (an agent's report, past its budget); a redirection to another stream
// is none.
func TestTheSplitterKeepsWrites(t *testing.T) {
	cases := []struct {
		line string
		d    shDialect
		want []string // op target, in order
	}{
		{"echo x > a.txt", shBash, []string{"> a.txt"}},
		{"echo x >> 'a b'.txt 2>&1", shBash, []string{">> a b.txt"}},
		{"cmd &> all.log", shBash, []string{"&> all.log"}},
		{"cmd >& all.log", shBash, []string{">& all.log"}},
		{"cmd >&2", shBash, nil},
		{"cmd 2>&-", shBash, nil},
		{"cmd < in.txt", shBash, nil},
		{"cmd <> rw.txt", shBash, []string{"<> rw.txt"}},
		{"cmd >| f", shBash, []string{">| f"}},
		{"cmd 3> f", shBash, []string{"3> f"}},
		{"{ echo x; } > g.txt", shBash, []string{"> g.txt"}},
		{"echo x *> p.txt", shPower, []string{"*> p.txt"}},
		{"echo x 2>&1", shPower, nil},
		{"echo x > c.txt", shCmd, []string{"> c.txt"}},
	}
	for _, c := range cases {
		l, why := shSplit(c.line, c.d, 0)
		if why != "" {
			t.Fatalf("%s: %s", c.line, why)
		}
		var got []string
		for _, cmd := range l.cmds {
			for _, r := range cmd.redirs {
				got = append(got, r.op+" "+r.target.text)
			}
		}
		if len(got) != len(c.want) {
			t.Errorf("%s: %q, want %q", c.line, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: %q, want %q", c.line, got, c.want)
			}
		}
	}
	for line, quoted := range map[string]bool{"cat <<'EOF'\nx\nEOF\n": true, "cat <<\"EOF\"\nx\nEOF\n": true, "cat <<EOF\nx\nEOF\n": false,
		"cat <<-E\\OF\n\tx\nEOF\n": true} {
		l, why := shSplit(line, shBash, 0)
		if why != "" || len(l.cmds) == 0 || len(l.cmds[0].feeds) != 1 || !l.cmds[0].feeds[0].doc || l.cmds[0].feeds[0].quoted != quoted {
			t.Errorf("%q: %v %+v", line, why, l)
		}
	}
}
