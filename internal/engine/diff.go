package engine

// update --diff: each file the plan writes, as a unified diff of its lines (three lines of context), in ASCII.

import (
	"fmt"
	"strings"
)

// maxDiffLines bounds the line-by-line comparison (its table grows as the product of the two files' lines).
const maxDiffLines = 3000

// Diff gives a unified diff from old to new, the files named a/<path> and b/<path>; "" when they are equal. A file
// with a NUL byte is binary, and a file past maxDiffLines lines is shown by its size only.
func Diff(path string, old, new []byte) string {
	if string(old) == string(new) {
		return ""
	}
	head := "--- a/" + path + "\n+++ b/" + path + "\n"
	if old == nil {
		head = "--- /dev/null\n+++ b/" + path + "\n"
	}
	if new == nil {
		head = "--- a/" + path + "\n+++ /dev/null\n"
	}
	if strings.ContainsRune(string(old), 0) || strings.ContainsRune(string(new), 0) {
		return head + "(binary: not shown)\n"
	}
	a, b := splitLines(old), splitLines(new)
	if len(a) > maxDiffLines || len(b) > maxDiffLines {
		return head + fmt.Sprintf("(too long to show: %d lines before, %d after)\n", len(a), len(b))
	}
	ops := lcsOps(a, b)
	var out strings.Builder
	out.WriteString(head)
	const ctx = 3
	// Hunks: each change with three lines around it; changes closer than six lines share a hunk.
	var changes []int
	for idx, o := range ops {
		if o.kind != ' ' {
			changes = append(changes, idx)
		}
	}
	for k := 0; k < len(changes); {
		start := changes[k] - ctx
		if start < 0 {
			start = 0
		}
		last := changes[k]
		k++
		for k < len(changes) && changes[k]-last <= 2*ctx {
			last = changes[k]
			k++
		}
		end := last + ctx + 1
		if end > len(ops) {
			end = len(ops)
		}
		aStart, bStart := ops[start].ai, ops[start].bi
		aLen, bLen := 0, 0
		for _, o := range ops[start:end] {
			if o.kind != '+' {
				aLen++
			}
			if o.kind != '-' {
				bLen++
			}
		}
		fmt.Fprintf(&out, "@@ -%s +%s @@\n", hunkRange(aStart, aLen), hunkRange(bStart, bLen))
		for _, o := range ops[start:end] {
			out.WriteString(string(o.kind) + ascii(strings.TrimSuffix(o.text, "\r")) + "\n")
		}
	}
	return out.String()
}

func hunkRange(start, n int) string {
	if n == 0 {
		return fmt.Sprintf("%d,0", start)
	}
	if n == 1 {
		return fmt.Sprintf("%d", start+1)
	}
	return fmt.Sprintf("%d,%d", start+1, n)
}

func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	s := strings.TrimSuffix(string(b), "\n")
	return strings.Split(s, "\n")
}

type diffOp struct {
	kind   byte // ' ', '-' or '+'
	text   string
	ai, bi int // the line's place before and after, for hunk headers
}

// lcsOps is the classic longest-common-subsequence diff.
func lcsOps(a, b []string) []diffOp {
	n, m := len(a), len(b)
	table := make([][]int32, n+1)
	for i := range table {
		table[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				table[i][j] = table[i+1][j+1] + 1
			case table[i+1][j] >= table[i][j+1]:
				table[i][j] = table[i+1][j]
			default:
				table[i][j] = table[i][j+1]
			}
		}
	}
	var ops []diffOp
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && a[i] == b[j]:
			ops = append(ops, diffOp{' ', a[i], i, j})
			i++
			j++
		case j < m && (i == n || table[i][j+1] > table[i+1][j]):
			ops = append(ops, diffOp{'+', b[j], i, j})
			j++
		default:
			ops = append(ops, diffOp{'-', a[i], i, j})
			i++
		}
	}
	return ops
}
