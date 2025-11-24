# Integration Guide

This guide provides instructions on how to integrate with the Bible API Service, using either the provided Go client library or direct HTTP requests.

## Authentication

All requests to the API must be authenticated using an API key. The API key should be included in the `X-API-KEY` header.

```bash
curl -X POST https://your-api-url.com/query \
  -H "X-API-KEY: your-api-key" \
  ...
```

## Using the Go Client

The Bible API Service provides a Go client library for easy integration.

### Installation

```bash
go get github.com/julwrites/BibleAIAPI/pkg/client
```

### Initialization

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/julwrites/BibleAIAPI/pkg/client"
)

func main() {
    // Initialize the client
    // Replace with your deployed URL and API Key
    c := client.NewClient("https://your-service-url.run.app", "your-api-key")

    // ... use the client
}
```

### Examples

#### Get Verse

```go
resp, err := c.GetVerses(context.Background(), []string{"John 3:16"}, "ESV")
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Verse)
```

#### Word Search

```go
resp, err := c.SearchWords(context.Background(), []string{"Grace"}, "ESV")
if err != nil {
    log.Fatal(err)
}
for _, result := range resp {
    fmt.Printf("%s: %s\n", result.Verse, result.URL)
}
```

#### Open Query

```go
resp, err := c.OpenQuery(context.Background(), "Who is Jesus?", "ESV")
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Text)
```

#### Chat with LLM

```go
schema := `{
    "type": "object",
    "properties": {
        "summary": {"type": "string"}
    }
}`
// Pass verses as context if needed
resp, err := c.Chat(context.Background(), "Summarize the verse", schema, []string{"John 3:16"}, "ESV")
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp["summary"])
```

## Direct API Usage

You can also interact with the API directly using HTTP requests.

### Get Verse

**Request:**

```bash
curl -X POST https://your-service-url.run.app/query \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: your-api-key" \
  -d '{
    "query": {
      "verses": ["John 3:16"]
    },
    "context": {
      "user": {
        "version": "ESV"
      }
    }
  }'
```

**Response:**

```json
{
  "verse": "John 3:16 (ESV) For God so loved the world..."
}
```

### Chat Query

**Request:**

```bash
curl -X POST https://your-service-url.run.app/query \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: your-api-key" \
  -d '{
    "query": {
      "prompt": "Explain this verse"
    },
    "context": {
      "schema": "{\"type\": \"object\", \"properties\": {\"explanation\": {\"type\": \"string\"}}}",
      "verses": ["John 3:16"],
      "user": {
        "version": "ESV"
      }
    }
  }'
```
