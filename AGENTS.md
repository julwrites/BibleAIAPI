# AI Agent Instructions

You are an expert Software Engineer working on this project. Your primary responsibility is to implement features and fixes while strictly adhering to the **Task Documentation System**.

## Core Philosophy
**"If it's not documented in `docs/tasks/`, it didn't happen."**

## Workflow
1.  **Pick a Task**: Run `python3 scripts/tasks.py next` to find the best task, `context` to see active tasks, or `list` to see pending ones.
2.  **Plan & Document**:
    *   **Memory Check**: Run `python3 scripts/memory.py list` (or use the Memory Skill) to recall relevant long-term information.
    *   **Security Check**: Ask the user about specific security considerations for this task.
    *   If starting a new task, use `scripts/tasks.py create` (or `python3 scripts/tasks.py create`) to generate a new task file.
    *   Update the task status: `python3 scripts/tasks.py update [TASK_ID] in_progress`.
3.  **Implement**: Write code, run tests.
4.  **Update Documentation Loop**:
    *   As you complete sub-tasks, check them off in the task document.
    *   If you hit a blocker, update status to `wip_blocked` and describe the issue in the file.
    *   Record key architectural decisions in the task document.
    *   **Memory Update**: If you learn something valuable for the long term, use `scripts/memory.py create` to record it.
5.  **Review & Verify**:
    *   Once implementation is complete, update status to `review_requested`: `python3 scripts/tasks.py update [TASK_ID] review_requested`.
    *   Ask a human or another agent to review the code.
    *   Once approved and tested, update status to `verified`.
6.  **Finalize**:
    *   Update status to `completed`: `python3 scripts/tasks.py update [TASK_ID] completed`.
    *   Record actual effort in the file.
    *   Ensure all acceptance criteria are met.

## Tools
*   **Wrapper**: `./scripts/tasks` (Checks for Python, recommended).
*   **Next**: `./scripts/tasks next` (Finds the best task to work on).
*   **Create**: `./scripts/tasks create [category] "Title"`
*   **List**: `./scripts/tasks list [--status pending]`
*   **Context**: `./scripts/tasks context`
*   **Update**: `./scripts/tasks update [ID] [status]`
*   **Migrate**: `./scripts/tasks migrate` (Migrate legacy tasks to new format)
*   **Memory**: `./scripts/memory.py [create|list|read]`
*   **JSON Output**: Add `--format json` to any command for machine parsing.

## Documentation Reference
*   **Guide**: Read `docs/tasks/GUIDE.md` for strict formatting and process rules.
*   **Architecture**: Refer to `docs/architecture/` for system design.
*   **Features**: Refer to `docs/features/` for feature specifications.
*   **Security**: Refer to `docs/security/` for risk assessments and mitigations.
*   **Memories**: Refer to `docs/memories/` for long-term project context.

## Code Style & Standards
*   Follow the existing patterns in the codebase.
*   Ensure all new code is covered by tests (if testing infrastructure exists).

## PR Review Methodology
When performing a PR review, follow this "Human-in-the-loop" process to ensure depth and efficiency.

### 1. Preparation
1.  **Create Task**: `python3 scripts/tasks.py create review "Review PR #<N>: <Title>"`
2.  **Fetch Details**: Use `gh` to get the PR context.
    *   `gh pr view <N>`
    *   `gh pr diff <N>`

### 2. Analysis & Planning (The "Review Plan")
**Do not review line-by-line yet.** Instead, analyze the changes and document a **Review Plan** in the task file (or present it for approval).

Your plan must include:
*   **High-Level Summary**: Purpose, new APIs, breaking changes.
*   **Dependency Check**: New libraries, maintenance status, security.
*   **Impact Assessment**: Effect on existing code/docs.
*   **Focus Areas**: Prioritized list of files/modules to check.
*   **Suggested Comments**: Draft comments for specific lines.
    *   Format: `File: <path> | Line: <N> | Comment: <suggestion>`
    *   Tone: Friendly, suggestion-based ("Consider...", "Nit: ...").

### 3. Execution
Once the human approves the plan and comments:
1.  **Pending Review**: Create a pending review using `gh`.
    *   `COMMIT_SHA=$(gh pr view <N> --json headRefOid -q .headRefOid)`
    *   `gh api repos/{owner}/{repo}/pulls/{N}/reviews -f commit_id="$COMMIT_SHA"`
2.  **Batch Comments**: Add comments to the pending review.
    *   `gh api repos/{owner}/{repo}/pulls/{N}/comments -f body="..." -f path="..." -f commit_id="$COMMIT_SHA" -F line=<L> -f side="RIGHT"`
3.  **Submit**:
    *   `gh pr review <N> --approve --body "Summary..."` (or `--request-changes`).

### 4. Close Task
*   Update task status to `completed`.

## Agent Interoperability
- **Task Manager Skill**: `.claude/skills/task_manager/`
- **Memory Skill**: `.claude/skills/memory/`
- **Tool Definitions**: `docs/interop/tool_definitions.json`


## Project Specific Instructions

## Overview

The Bible API Service is a Go-based stateless microservice designed for serverless deployment (Google Cloud Run). It acts as a bridge between clients, the Bible Gateway website (via scraping), and Large Language Models (LLMs).

## Key Technologies & Architecture

-   **Go**: Primary language (Go 1.24+).
-   **Docker**: Multi-stage builds for optimized container images.
-   **Scraping**: Uses `goquery` to scrape `classic.biblegateway.com`. Logic distinguishes between **Prose** and **Poetry** to preserve formatting.
-   **LLM Clients**: Modular architecture in `internal/llm`. Supports multiple providers (OpenAI, Gemini, DeepSeek, OpenRouter, custom OpenAI-compatible endpoints).
    -   **Fallback**: The `FallbackClient` (`internal/llm/client.go`) attempts providers in order defined by `LLM_CONFIG` JSON (or deprecated `LLM_PROVIDERS` env var).
-   **Feature Flags**: Managed via `go-feature-flag`.
    -   **Source**: Primarily loaded from the GitHub repository (`julwrites/BibleAIAPI`).
    -   **Fallback**: Falls back to a local file (`configs/flags.yaml`) if GitHub retrieval fails.
-   **Secrets**: Managed via `internal/secrets`.
    -   **Source**: Attempts to fetch from Google Secret Manager first.
    -   **Fallback**: Falls back to environment variables (useful for local development).

## Project Structure

-   `cmd/server`: Application entry point.
-   `internal/biblegateway`: Scraper logic. **Critical**: Must handle `div.poetry` for poetry preservation.
-   `internal/handlers`: HTTP handlers. `QueryHandler` routes requests based on payload (`verses` vs `words` vs `prompt`).
    -   **Validation**: The API enforces strict validation. A query must contain exactly one type (`verses`, `words`, `prompt`). The `context` object is only permitted when `query.prompt` is present.
-   `internal/llm`: Modular LLM provider implementations.
-   `internal/middleware`: Auth middleware (`X-API-KEY`) and logging.
-   `internal/secrets`: Secret retrieval logic.
-   `configs`: Local fallback configuration files.
-   `docs`: Detailed documentation (API, Architecture, Deployment).

## Scraper Implementation Details

The scraper (`internal/biblegateway/scraper.go`) is sensitive to HTML structure.
-   **Allowed Tags**: `h1`, `h2`, `h3`, `h4`, `p`, `span`, `i`, `br`, `sup` (for verse numbers).
-   **Sanitization**: Aggressive cleanup for prose; structure preservation for poetry.
-   **Non-breaking Spaces**: Must be converted to regular spaces (`\u00a0` -> ` `).

## Development Workflow

1.  **Understand the Goal**: Read `README.md` and `docs/` first.
2.  **Verify State**: Always check `go.mod` and existing code before assuming dependencies or logic.
3.  **Testing**:
    -   **Unit Tests**: Write table-driven tests. Use `go test ./...`.
    -   **Coverage**: Strict ratcheting coverage is enforced. Coverage must typically exceed 80%. Use `go tool cover` to check.
    -   **Scraper Tests**: Use realistic HTML snippets in tests, not mock strings.
    -   **Integration Tests**: Run `main` in a goroutine and make real HTTP requests (mocking external calls if needed).
4.  **Documentation**: Update `docs/` and `README.md` if logic changes.
5.  **Pre-Commit**: Ensure all tests pass and coverage is sufficient before submitting.

## Deployment

-   Refer to `docs/deployment.md` for detailed instructions.
-   **Secrets**: `API_KEY`, `GCP_PROJECT_ID`, and relevant LLM keys (`OPENAI_API_KEY`, etc.) must be set.
