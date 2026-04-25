# Build Coding Agents From Scratch in Go

![learnwithparam.com](https://www.learnwithparam.com/ai-bootcamp/opengraph-image)

Build a small, local coding agent in pure Go that reads files, writes files, and runs shell commands on your behalf. No SDK, no framework, just the Go standard library talking to an OpenAI-compatible LLM endpoint.

> Start learning at [learnwithparam.com](https://learnwithparam.com). Regional pricing available with discounts of up to 60%.

## What You'll Learn

- How a tool-calling loop actually works, with zero SDK magic
- Designing a clean tool interface (name, JSON schema, runner function)
- Wiring Go structs to the OpenAI-compatible `/chat/completions` contract
- Safely bridging an LLM to the filesystem and shell with a denylist
- Running a static, distroless Go binary as a portable agent

## Tech Stack

- **Go 1.22** - Single-binary runtime, no external dependencies
- **Go standard library only** - `net/http`, `encoding/json`, `os/exec`
- **OpenRouter** - OpenAI-compatible endpoint that fronts many open and closed models
- **Distroless Docker image** - Minimal, static, nonroot runtime
- **Make** - One-command setup, run, test, and container targets

## Getting Started

### Prerequisites

- Go 1.22 or newer
- An [OpenRouter API key](https://openrouter.ai)
- Docker (optional, for the container workflow)

### Quick Start

```bash
make setup          # Copy .env.example to .env and tidy modules
# Edit .env and paste your OpenRouter key
make run            # Start the REPL
```

Ask it something like:

```
> read the go.mod file and tell me the module path
```

### Piped Usage

The binary also reads a single prompt from stdin, which is useful for scripts and smoke tests:

```bash
echo "list the files in the agent directory" | OPENROUTER_API_KEY=sk-... go run .
```

### With Docker

```bash
make docker-build   # Build the image
make up             # Start the container (binds the repo as /workspace)
make down           # Stop it
```

## Challenges

Work through these incrementally to build the full agent:

1. **Echo the LLM** - Wire up `agent/llm.go` and confirm a plain chat round-trip works before tools enter the picture.
2. **Register read_file** - Define the first tool spec and make the model call it to answer a question about a file on disk.
3. **Close the Loop** - Execute tool calls, append their outputs as `role: tool` messages, and resend until the model emits a final answer.
4. **Add write_file** - Extend the tool set so the agent can create or overwrite files inside the working directory.
5. **Add run_bash with guardrails** - Shell execution with a denylist and a hard timeout so runaway commands cannot take down your machine.
6. **Cap the Iterations** - Enforce a `MaxIters` ceiling and return a clean error when the loop refuses to converge.
7. **Containerize It** - Produce a static binary, drop it into a distroless image, and run the agent with the repo mounted as `/workspace`.

## Makefile Targets

```
make help           Show all available commands
make setup          Copy .env and tidy Go modules
make run            Start the agent REPL
make build          Build a local binary at bin/agent
make test           Run unit tests
make docker-build   Build the Docker image
make up             Start the container via docker compose
make down           Stop the container
make clean          Remove build artifacts
```

## Learn more

- Start the course: [learnwithparam.com/courses/coding-agents-from-scratch-go](https://www.learnwithparam.com/courses/coding-agents-from-scratch-go)
- AI Bootcamp for Software Engineers: [learnwithparam.com/ai-bootcamp](https://www.learnwithparam.com/ai-bootcamp)
- All courses: [learnwithparam.com/courses](https://www.learnwithparam.com/courses)
