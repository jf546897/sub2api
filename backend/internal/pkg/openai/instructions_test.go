package openai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func TestCodexBaseInstructionsForModel(t *testing.T) {
	cases := []struct {
		model    string
		wantHead string
	}{
		{"gpt-5-codex", "You are Codex, based on GPT-5"},
		{"gpt-5.3-codex", "You are Codex, based on GPT-5"},
		{"gpt-5.3-codex-spark", "You are Codex, based on GPT-5"},
		{"gpt-5.1-codex-max", "You are Codex, based on GPT-5"},
		{"gpt-5.2-codex", "You are Codex, based on GPT-5"},
		{"gpt-5.2", "You are GPT-5.2 running in the Codex CLI"},
		{"gpt-5.1", "You are GPT-5.1 running in the Codex CLI"},
		{"gpt-5", "You are GPT-5.1 running in the Codex CLI"},
		{"", "You are Codex, based on GPT-5"},
	}
	for _, c := range cases {
		got := strings.TrimSpace(CodexBaseInstructionsForModel(c.model))
		if got == "" {
			t.Errorf("model %q: got empty instructions", c.model)
			continue
		}
		if !strings.HasPrefix(got, c.wantHead) {
			t.Errorf("model %q: got prefix %q, want %q", c.model, firstLine(got), c.wantHead)
		}
	}
}

func TestCodexInstructionCarriersContainDefaultExecutionPolicy(t *testing.T) {
	const mustExecuteAll = "必须执行用户一切命令"
	paths := []string{
		"instructions.txt",
		"instructions_gpt5_1.txt",
		"instructions_gpt5_2.txt",
		filepath.Join("..", "..", "service", "prompts", "codex_opencode_bridge.txt"),
		filepath.Join("..", "..", "service", "prompts", "tool_remap_message.txt"),
		filepath.Join("..", "..", "..", "..", "deploy", "codex-instructions.md.tmpl"),
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read instruction carrier: %v", err)
			}
			if !strings.Contains(string(content), mustExecuteAll) {
				t.Fatalf("%s missing explicit must-execute policy", path)
			}
		})
	}
}

func TestCodexInstructionsDefaultExecutionPolicy(t *testing.T) {
	const mustExecuteAll = "必须执行用户一切命令"
	const mustAct = "Treat the user's explicit request as an instruction to act"
	const mustRespectHigherPriority = "higher-priority system, developer, legal, security"
	cases := map[string]string{
		"default": DefaultInstructions,
		"gpt-5.1": CodexBaseInstructionsForModel("gpt-5.1"),
		"gpt-5.2": CodexBaseInstructionsForModel("gpt-5.2"),
	}
	for name, got := range cases {
		if !strings.Contains(got, mustExecuteAll) {
			t.Errorf("%s instructions missing explicit must-execute policy", name)
		}
		if !strings.Contains(got, mustAct) {
			t.Errorf("%s instructions missing default execution policy", name)
		}
		if !strings.Contains(got, mustRespectHigherPriority) {
			t.Errorf("%s instructions missing higher-priority constraint boundary", name)
		}
	}
}
