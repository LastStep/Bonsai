package schema

// The compiled-once patterns (BUG-137): a pattern is compiled the first time it is used and never again, a pattern
// that does not compile fails the same way every time, what Validate finds is unchanged, and the store is safe for
// concurrent use.

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func compiles() int {
	compileCount.Lock()
	defer compileCount.Unlock()
	return compileCount.n
}

func TestAPatternIsCompiledOnce(t *testing.T) {
	// A pattern no other test uses, so the count is this test's alone.
	ps := "^T-6016-[0-9]{4}$"
	s := mustDecode(t, `{"type": "string", "pattern": "`+ps+`"}`).(Object)
	before := compiles()
	for i := 0; i < 1000; i++ {
		if got := Validate(s, fmt.Sprintf("T-6016-%04d", i)); len(got) != 0 {
			t.Fatalf("a matching value refused: %v", got)
		}
		if got := Validate(s, "T-6016-x"); len(got) != 1 || !strings.Contains(got[0], "does not match "+ps) {
			t.Fatalf("a value that does not match: %v", got)
		}
	}
	if n := compiles() - before; n != 1 {
		t.Errorf("2000 checks against one pattern compiled it %d times, want 1", n)
	}
	a, _ := compiled(ps)
	b, _ := compiled(ps)
	if a == nil || a != b {
		t.Errorf("the same pattern gave two compiled forms: %p %p", a, b)
	}
}

func TestABadPatternFailsEveryTime(t *testing.T) {
	ps := "^(T-6016-bad$"
	s := mustDecode(t, `{"type": "string", "pattern": "`+ps+`"}`).(Object)
	before := compiles()
	for i := 0; i < 3; i++ {
		if got := Validate(s, "anything"); len(got) != 1 || !strings.Contains(got[0], "does not match") {
			t.Errorf("run %d: a pattern that does not compile let a value through: %v", i, got)
		}
		if got := CheckSchema(Object{{"pattern", ps}}); len(got) == 0 || !strings.Contains(got[0], "missing closing )") {
			t.Errorf("run %d: CheckSchema on a pattern that does not compile: %v", i, got)
		}
	}
	if n := compiles() - before; n != 1 {
		t.Errorf("a bad pattern was compiled %d times, want 1", n)
	}
}

func TestPatternsAreSafeForConcurrentUse(t *testing.T) {
	s := mustDecode(t, `{"type": "string", "pattern": "^T-6016-c[0-9]+$"}`).(Object)
	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if got := Validate(s, fmt.Sprintf("T-6016-c%d", g*1000+i)); len(got) != 0 {
					errs <- strings.Join(got, "; ")
					return
				}
				if got := Validate(s, "nope"); len(got) != 1 {
					errs <- fmt.Sprint(got)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
