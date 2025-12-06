# Instruction-Driven Workflows Plan

This document outlines the plan to implement instruction-driven workflows in the Bible API Service. The goal is to move from a single hardcoded chat flow to a flexible, strategy-based architecture where the `context.command` field determines the processing logic (workflow).

## 1. API Design Changes

### 1.1 Update `QueryRequest`
Modify `internal/handlers/types.go` to include a `command` field in the `context` object.

```go
type QueryRequest struct {
    // ... existing fields
    Context struct {
        // ... existing fields
        Command string `json:"command,omitempty"` // New field: "bible_query", "word_study", "bible_study_plan"
    } `json:"context,omitempty"`
}
```

*   **Default Behavior**: If `command` is empty, default to `"bible_query"`.
*   **Validation**: Ensure `command` is one of the allowed values.

### 1.2 Update `LLMClient` Interface
Refactor `internal/llm/provider/provider.go` to support system prompts, which are essential for instruction-driven workflows.

```go
type LLMClient interface {
    // Changed: Added systemPrompt parameter
    Query(ctx context.Context, systemPrompt string, userPrompt string, schema string) (string, error)
}
```

*   **Impact**: This requires updating all implementations in `internal/llm/` (openai, gemini, deepseek, openapicustom) and the `FallbackClient`.

## 2. Architecture Refactor

### 2.1 Workflow Strategy Pattern
Refactor `ChatService` (`internal/chat/chat.go`) to use a strategy pattern.

**New Interface:**
```go
type Workflow interface {
    Execute(ctx context.Context, req Request, clients Clients) (Response, error)
}

type Clients struct {
    BibleGateway BibleGatewayClient
    LLM          provider.LLMClient
}
```

**ChatService Update:**
The `ChatService` will now act as a factory/dispatcher:
1.  Determine the workflow based on `req.Command`.
2.  Instantiate the appropriate `Workflow` implementation.
3.  Call `Execute`.

## 3. Workflows

### 3.1 Bible Query (`bible_query`)
*   **Goal**: Standard Q&A about the Bible.
*   **Logic**:
    1.  **Reference Extraction** (New): If `req.VerseRefs` is empty, use an LLM call (or robust regex) to extract verse references from `req.Prompt`.
        *   *Note*: This fills the gap identified in the current legacy workflow.
    2.  **Retrieve Verses**: Fetch verses using `BibleGatewayClient`.
    3.  **LLM Query**:
        *   **System Prompt**: "You are a helpful Bible assistant. strict checks: 1. Input must be about the Bible. 2. Input must have at least one verse reference (or we found one). 3. Input must be sanitized. If checks fail, return a refusal message."
        *   **User Prompt**: The user's prompt + retrieved verses.

### 3.2 Word Study (`word_study`)
*   **Goal**: Deep dive into specific words/phrases.
*   **Logic**:
    1.  **Keyword Extraction**: Call LLM to break down the user's prompt into key words/phrases to study.
    2.  **Word Search**: Use existing `BibleGatewayClient.SearchWords` to find occurrences.
    3.  **Definition Retrieval** (New):
        *   Implement `BibleGatewayClient.GetDefinition(word)`.
        *   *Implementation Detail*: Scrape `biblegateway.com/quicksearch/?quicksearch={word}` and parse the "Topical" or "Dictionary" results (e.g., Nave's Topical Bible).
    4.  **LLM Query**:
        *   **System Prompt**: "Provide a detailed word study. Include definitions, occurrences, and theological context."
        *   **User Prompt**: User prompt + Search Results + Definitions.

### 3.3 Bible Study Plan (`bible_study_plan`)
*   **Goal**: Generate a structured study guide.
*   **Logic**:
    1.  **Plan Generation**: Call LLM to identify relevant passages and key themes/words based on the user's topic.
    2.  **Data Retrieval**: Fetch identified verses and definitions.
    3.  **LLM Query**:
        *   **System Prompt**: "Create a Bible study plan. You MUST include the following sections: Passage(s) to be studied, Discussion questions, Application questions."
        *   **User Prompt**: User topic + Retrieved Context.

## 4. Implementation Tasks

1.  **Core Refactoring**:
    *   Update `LLMClient` interface and all providers.
    *   Update `QueryRequest` struct.
    *   Refactor `ChatService` to support `Workflow` interface.
2.  **Scraper Updates**:
    *   Implement `GetDefinition` in `BibleGatewayClient` (likely parsing "Topical" results from QuickSearch).
3.  **Workflow Implementation**:
    *   Implement `BibleQueryWorkflow`.
    *   Implement `WordStudyWorkflow`.
    *   Implement `BibleStudyPlanWorkflow`.
4.  **Testing**:
    *   Unit tests for each workflow.
    *   Integration tests for the new API parameters.

## 5. Documentation
*   Update `docs/api_design.md` to reflect the new `context.command` field.
