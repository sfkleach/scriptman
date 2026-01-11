package interpreter_test

import (
	"strings"
	"testing"

	"github.com/sfkleach/scriptman/pkg/interpreter"
)

// DetermineInterpreterChoices is a test helper that wraps the DecisionInput method.
// This function is only used in tests to avoid exposing the standalone function in the public API.
func DetermineInterpreterChoices(scriptPath string, scriptContent []byte, explicitInterpreter string, trustShebang bool) interpreter.DecisionResult {
	input := interpreter.NewDecisionInput(scriptPath, scriptContent, explicitInterpreter, trustShebang)
	return input.DetermineInterpreterChoices()
}

func TestDetermineInterpreterChoices_ExplicitInterpreter(t *testing.T) {
	content := []byte("#!/usr/bin/env python\nprint('hello')")
	result := DetermineInterpreterChoices("script.py", content, "ruby", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "explicit" {
		t.Errorf("expected source 'explicit', got %s", choice.Source)
	}
	if choice.Interpreter != "ruby" {
		t.Errorf("expected interpreter 'ruby', got %s", choice.Interpreter)
	}
	if choice.RequiresPrompt {
		t.Errorf("explicit interpreter should not require prompt")
	}
}

func TestDetermineInterpreterChoices_ConsistentShebangAndExtension(t *testing.T) {
	content := []byte("#!/usr/bin/env python\nprint('hello')")
	result := DetermineInterpreterChoices("script.py", content, "", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "extension-alternatives" {
		t.Errorf("expected source 'extension-alternatives', got %s", choice.Source)
	}
	if choice.RequiresPrompt {
		t.Errorf("consistent shebang+extension should not require prompt")
	}
}

func TestDetermineInterpreterChoices_InconsistentShebangAndExtension(t *testing.T) {
	content := []byte("#!/usr/bin/env ruby\nputs 'hello'")
	result := DetermineInterpreterChoices("script.py", content, "", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(result.Choices))
	}

	// Both choices should require prompt.
	for i, choice := range result.Choices {
		if !choice.RequiresPrompt {
			t.Errorf("choice %d should require prompt", i)
		}
	}

	// First choice should be extension-based (recommended).
	if result.Choices[0].Source != "extension-alternatives" {
		t.Errorf("first choice should be extension-based, got %s", result.Choices[0].Source)
	}

	// Second choice should be shebang-based.
	if result.Choices[1].Source != "shebang" {
		t.Errorf("second choice should be shebang-based, got %s", result.Choices[1].Source)
	}
}

func TestDetermineInterpreterChoices_ShebangWithArguments(t *testing.T) {
	content := []byte("#!/usr/bin/python -u\nprint('hello')")
	result := DetermineInterpreterChoices("script.py", content, "", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 2 {
		t.Fatalf("expected 2 choices (args present), got %d", len(result.Choices))
	}

	// Both choices should require prompt.
	for i, choice := range result.Choices {
		if !choice.RequiresPrompt {
			t.Errorf("choice %d should require prompt", i)
		}
	}
}

func TestDetermineInterpreterChoices_NoExtension(t *testing.T) {
	content := []byte("#!/usr/bin/env python3\nprint('hello')")
	result := DetermineInterpreterChoices("script", content, "", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "shebang" {
		t.Errorf("expected source 'shebang', got %s", choice.Source)
	}

	// Using env form should not require prompt.
	if choice.RequiresPrompt {
		t.Errorf("env form without extension should not require prompt")
	}
}

func TestDetermineInterpreterChoices_NoShebangWithExtension(t *testing.T) {
	content := []byte("print('hello')")
	result := DetermineInterpreterChoices("script.py", content, "", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "extension-alternatives" {
		t.Errorf("expected source 'extension-alternatives', got %s", choice.Source)
	}
	if choice.RequiresPrompt {
		t.Errorf("extension-only should not require prompt")
	}
}

func TestDetermineInterpreterChoices_NoShebangNoExtension(t *testing.T) {
	content := []byte("print('hello')")
	result := DetermineInterpreterChoices("script", content, "", false)

	if result.Error == nil {
		t.Fatal("expected error for no shebang and no extension")
	}

	if len(result.Choices) != 0 {
		t.Errorf("expected 0 choices for error case, got %d", len(result.Choices))
	}
}

func TestDetermineInterpreterChoices_UnrecognizedExtension(t *testing.T) {
	content := []byte("#!/usr/bin/env foo\nsome code")
	result := DetermineInterpreterChoices("script.xyz", content, "", false)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "shebang" {
		t.Errorf("expected source 'shebang', got %s", choice.Source)
	}
	if !choice.RequiresPrompt {
		t.Errorf("unrecognized extension should require prompt")
	}
}

func TestDetermineInterpreterChoices_TrustShebang(t *testing.T) {
	// Test that --trust-shebang bypasses all consistency checks.

	// Case 1: Inconsistent shebang + extension (normally prompts, but trusts shebang).
	content := []byte("#!/usr/bin/env ruby\nputs 'hello'")
	result := DetermineInterpreterChoices("script.py", content, "", true)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice with trust-shebang, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "shebang" {
		t.Errorf("expected source 'shebang', got %s", choice.Source)
	}
	if choice.Interpreter != "ruby" {
		t.Errorf("expected interpreter 'ruby', got %s", choice.Interpreter)
	}
	if choice.RequiresPrompt {
		t.Errorf("trust-shebang should not require prompt")
	}
	if !strings.Contains(choice.Reason, "trust-shebang") {
		t.Errorf("reason should mention trust-shebang flag: %s", choice.Reason)
	}
}

func TestDetermineInterpreterChoices_TrustShebangWithArguments(t *testing.T) {
	// Even with arguments, --trust-shebang should bypass prompts.
	content := []byte("#!/usr/bin/python -u -W ignore\nprint('hello')")
	result := DetermineInterpreterChoices("script.py", content, "", true)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("expected 1 choice with trust-shebang, got %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.Source != "shebang" {
		t.Errorf("expected source 'shebang', got %s", choice.Source)
	}
	if choice.RequiresPrompt {
		t.Errorf("trust-shebang should bypass prompts even with arguments")
	}
}
