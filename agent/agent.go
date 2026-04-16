package agent

import (
	"fmt"
	"strings"
)

// Agent drives a tool-calling loop against an OpenAI-compatible LLM.
type Agent struct {
	LLM      LLMClient
	Tools    []Tool
	MaxIters int
	System   string
}

// New constructs an Agent with the default builtin tools.
func New(llm LLMClient, maxIters int) *Agent {
	if maxIters <= 0 {
		maxIters = 10
	}
	return &Agent{
		LLM:      llm,
		Tools:    BuiltinTools(),
		MaxIters: maxIters,
		System: strings.TrimSpace(`
You are a small coding agent running locally on the user's machine.
You have three tools: read_file, write_file, run_bash.
Prefer reading before writing. When the user asks a direct question that can
be answered from a file, read the file and answer plainly. Be concise.
`),
	}
}

// findTool looks up a tool by function name.
func (a *Agent) findTool(name string) *Tool {
	for i := range a.Tools {
		if a.Tools[i].Spec.Function.Name == name {
			return &a.Tools[i]
		}
	}
	return nil
}

// specs returns the tool specs in the wire format.
func (a *Agent) specs() []ToolSpec {
	out := make([]ToolSpec, 0, len(a.Tools))
	for _, t := range a.Tools {
		out = append(out, t.Spec)
	}
	return out
}

// Run executes a single user turn with the tool-calling loop.
func (a *Agent) Run(userMsg string) (string, error) {
	messages := []Message{
		{Role: "system", Content: a.System},
		{Role: "user", Content: userMsg},
	}

	for iter := 0; iter < a.MaxIters; iter++ {
		resp, err := a.LLM.Chat(ChatRequest{
			Messages: messages,
			Tools:    a.specs(),
		})
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no choices in response")
		}
		choice := resp.Choices[0]
		messages = append(messages, choice.Message)

		if len(choice.Message.ToolCalls) > 0 {
			for _, tc := range choice.Message.ToolCalls {
				tool := a.findTool(tc.Function.Name)
				var result string
				if tool == nil {
					result = fmt.Sprintf("error: unknown tool %q", tc.Function.Name)
				} else {
					out, terr := tool.Run(tc.Function.Arguments)
					if terr != nil {
						result = fmt.Sprintf("error: %v\n%s", terr, out)
					} else {
						result = out
					}
				}
				messages = append(messages, Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Content:    result,
				})
			}
			continue
		}

		if choice.FinishReason == "stop" || choice.Message.Content != "" {
			return strings.TrimSpace(choice.Message.Content), nil
		}
	}
	return "", fmt.Errorf("agent: exceeded max iterations (%d)", a.MaxIters)
}
