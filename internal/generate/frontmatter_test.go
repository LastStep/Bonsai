package generate

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	input := []byte(`---
description: End-of-session verification and cleanup
display_name: Session Wrap-Up
---

# Session Wrap-Up
Content here.
`)
	meta, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if meta.Description != "End-of-session verification and cleanup" {
		t.Errorf("description = %q, want %q", meta.Description, "End-of-session verification and cleanup")
	}
	if meta.DisplayName != "Session Wrap-Up" {
		t.Errorf("display_name = %q, want %q", meta.DisplayName, "Session Wrap-Up")
	}
}

func TestParseFrontmatterSensor(t *testing.T) {
	input := []byte(`---
description: Blocks risky operations on Fridays
event: PreToolUse
matcher: Bash
---

#!/usr/bin/env bash
echo "hello"
`)
	meta, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if meta.Event != "PreToolUse" {
		t.Errorf("event = %q, want %q", meta.Event, "PreToolUse")
	}
	if meta.Matcher != "Bash" {
		t.Errorf("matcher = %q, want %q", meta.Matcher, "Bash")
	}
}

func TestParseFrontmatterRoutine(t *testing.T) {
	input := []byte(`---
description: Clean up stale branches
frequency: 14 days
---

# Branch Cleanup
`)
	meta, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if meta.Frequency != "14 days" {
		t.Errorf("frequency = %q, want %q", meta.Frequency, "14 days")
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	input := []byte(`# Just a markdown file
No frontmatter here.
`)
	_, err := ParseFrontmatter(input)
	if err == nil {
		t.Error("expected error for file without frontmatter")
	}
}

func TestParseFrontmatterUnterminated(t *testing.T) {
	input := []byte(`---
description: Oops no closing delimiter
`)
	_, err := ParseFrontmatter(input)
	if err == nil {
		t.Error("expected error for unterminated frontmatter")
	}
}

func TestParseFrontmatterExtraFields(t *testing.T) {
	// Extra YAML fields like "tags" should be ignored gracefully
	input := []byte(`---
tags: [workflow, session]
description: End-of-session verification
---

# Content
`)
	meta, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if meta.Description != "End-of-session verification" {
		t.Errorf("description = %q", meta.Description)
	}
}

func TestParseFrontmatterMinimal(t *testing.T) {
	input := []byte(`---
description: Just a description
---
`)
	meta, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if meta.Description != "Just a description" {
		t.Errorf("description = %q", meta.Description)
	}
	if meta.DisplayName != "" {
		t.Errorf("display_name should be empty, got %q", meta.DisplayName)
	}
}
