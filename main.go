package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/learnwithparam/coding-agents-from-scratch-go/agent"
)

func main() {
	provider := getenvDefault("LLM_PROVIDER", "openrouter")
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	model := getenvDefault("OPENROUTER_MODEL", "google/gemma-3-12b-it")

	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: OPENROUTER_API_KEY is not set. Copy .env.example to .env and export it.")
		os.Exit(1)
	}
	if provider != "openrouter" {
		fmt.Fprintf(os.Stderr, "warning: LLM_PROVIDER=%q is not supported yet. Falling back to openrouter.\n", provider)
	}

	llm := agent.NewOpenRouterClient(apiKey, model)
	a := agent.New(llm, 10)

	fmt.Println("coding-agents-from-scratch-go")
	fmt.Println("Type a request, or /exit to quit.")

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// Support non-interactive piped input: read all lines joined.
	stat, _ := os.Stdin.Stat()
	piped := (stat.Mode() & os.ModeCharDevice) == 0

	if piped {
		var lines []string
		for in.Scan() {
			lines = append(lines, in.Text())
		}
		msg := strings.TrimSpace(strings.Join(lines, "\n"))
		if msg == "" {
			return
		}
		out, err := a.Run(msg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	for {
		fmt.Print("\n> ")
		if !in.Scan() {
			return
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		if line == "/exit" || line == "/quit" {
			return
		}
		out, err := a.Run(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}
		fmt.Println(out)
	}
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
