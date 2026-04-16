package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileTool(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(fp, []byte("hello coding agent"), 0o644); err != nil {
		t.Fatal(err)
	}
	args, _ := json.Marshal(map[string]string{"path": fp})
	out, err := runReadFile(string(args))
	if err != nil {
		t.Fatalf("runReadFile: %v", err)
	}
	if !strings.Contains(out, "hello coding agent") {
		t.Fatalf("unexpected content: %q", out)
	}
}

func TestWriteFileTool(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "sub", "out.txt")
	args, _ := json.Marshal(map[string]string{"path": fp, "content": "abc"})
	if _, err := runWriteFile(string(args)); err != nil {
		t.Fatalf("runWriteFile: %v", err)
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "abc" {
		t.Fatalf("expected abc, got %q", string(data))
	}
}

func TestRunBashDenylist(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"cmd": "rm -rf /"})
	_, err := runBash(string(args))
	if err == nil {
		t.Fatal("expected denylist rejection")
	}
}

func TestAgentBuiltinToolsRegistered(t *testing.T) {
	a := New(nil, 3)
	want := []string{"read_file", "write_file", "run_bash"}
	for _, name := range want {
		if a.findTool(name) == nil {
			t.Errorf("missing tool %s", name)
		}
	}
}
