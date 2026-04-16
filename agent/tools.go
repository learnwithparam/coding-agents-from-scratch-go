package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Tool is a runnable capability the agent can invoke.
type Tool struct {
	Spec ToolSpec
	Run  func(args string) (string, error)
}

// BuiltinTools returns the default tool set (read_file, write_file, run_bash).
func BuiltinTools() []Tool {
	return []Tool{
		{
			Spec: ToolSpec{
				Type: "function",
				Function: FunctionSpec{
					Name:        "read_file",
					Description: "Read a UTF-8 text file from the local working directory and return its contents.",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"path": map[string]any{
								"type":        "string",
								"description": "Relative path to the file to read.",
							},
						},
						"required": []string{"path"},
					},
				},
			},
			Run: runReadFile,
		},
		{
			Spec: ToolSpec{
				Type: "function",
				Function: FunctionSpec{
					Name:        "write_file",
					Description: "Create or overwrite a file with the given text content.",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"path":    map[string]any{"type": "string", "description": "Relative path."},
							"content": map[string]any{"type": "string", "description": "Full file content."},
						},
						"required": []string{"path", "content"},
					},
				},
			},
			Run: runWriteFile,
		},
		{
			Spec: ToolSpec{
				Type: "function",
				Function: FunctionSpec{
					Name:        "run_bash",
					Description: "Run a shell command and return combined stdout+stderr. Dangerous commands are blocked.",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"cmd": map[string]any{"type": "string", "description": "Command line to execute with /bin/sh -c."},
						},
						"required": []string{"cmd"},
					},
				},
			},
			Run: runBash,
		},
	}
}

func runReadFile(args string) (string, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "", fmt.Errorf("read_file: invalid args: %w", err)
	}
	if in.Path == "" {
		return "", fmt.Errorf("read_file: path is required")
	}
	clean := filepath.Clean(in.Path)
	data, err := os.ReadFile(clean)
	if err != nil {
		return "", fmt.Errorf("read_file(%s): %w", clean, err)
	}
	return string(data), nil
}

func runWriteFile(args string) (string, error) {
	var in struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "", fmt.Errorf("write_file: invalid args: %w", err)
	}
	if in.Path == "" {
		return "", fmt.Errorf("write_file: path is required")
	}
	clean := filepath.Clean(in.Path)
	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return "", fmt.Errorf("write_file mkdir: %w", err)
	}
	if err := os.WriteFile(clean, []byte(in.Content), 0o644); err != nil {
		return "", fmt.Errorf("write_file: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(in.Content), clean), nil
}

// denylist keeps obviously destructive shell fragments out of the loop.
var denylist = []string{
	"rm -rf /", "mkfs", ":(){:|:&};:", "shutdown", "reboot", "dd if=", "> /dev/sda",
}

func runBash(args string) (string, error) {
	var in struct {
		Cmd string `json:"cmd"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "", fmt.Errorf("run_bash: invalid args: %w", err)
	}
	if strings.TrimSpace(in.Cmd) == "" {
		return "", fmt.Errorf("run_bash: cmd is required")
	}
	for _, bad := range denylist {
		if strings.Contains(in.Cmd, bad) {
			return "", fmt.Errorf("run_bash: refused to run dangerous command containing %q", bad)
		}
	}
	cmd := exec.Command("/bin/sh", "-c", in.Cmd)
	cmd.Env = os.Environ()
	// Soft timeout via goroutine; keep dependency-free.
	done := make(chan struct{})
	var out []byte
	var err error
	go func() {
		out, err = cmd.CombinedOutput()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		return "", fmt.Errorf("run_bash: timed out after 30s")
	}
	if err != nil {
		return string(out), fmt.Errorf("run_bash exit: %w", err)
	}
	return string(out), nil
}
