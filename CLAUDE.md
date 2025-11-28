# AI-Flow Developer Context Guide

This document provides comprehensive context for AI coding assistants (like Claude) to understand the ai-flow codebase architecture, design patterns, and conventions. Use this guide when adding new features to maintain project consistency.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture & Design Patterns](#architecture--design-patterns)
3. [Package Structure](#package-structure)
4. [Adding New LLM Providers](#adding-new-llm-providers)
5. [Testing Conventions](#testing-conventions)
6. [Example Patterns](#example-patterns)
7. [Configuration Guidelines](#configuration-guidelines)
8. [Code Organization Best Practices](#code-organization-best-practices)

---

## Project Overview

**ai-flow** is a Go-based AI abstraction layer that provides unified interfaces for working with various AI services and building AI-powered applications.

### Core Functionality

- **LLM Clients**: OpenAI and Ollama integrations with unified interface
- **RAG (Retrieval-Augmented Generation)**: v1 and v2 implementations with retriever/synthesizer patterns
- **Vector Databases**: Abstractions over chromem-go with simple in-memory fallback
- **Text Splitting**: Recursive text chunking with multiple tokenizer strategies (whitespace, TikToken)
- **Utilities**: dotenv loader, HTTP client wrappers

### Key Dependencies

```go
github.com/philippgille/chromem-go v0.7.0      // Vector database
github.com/sashabaranov/go-openai v1.41.2      // OpenAI client
github.com/neurosnap/sentences v1.1.2          // Sentence segmentation
github.com/pkoukk/tiktoken-go v0.1.8           // Token counting
github.com/stretchr/testify v1.11.1            // Testing framework
```

---

## Architecture & Design Patterns

Understanding these patterns is critical for maintaining consistency when adding features.

### 1. Interface-Driven Design

All major components define interfaces first, implementations second.

**Example**: `llm/iface/iface.go`
```go
package iface

type LLM interface {
    ListModels(ctx context.Context) ([]*models.Model, error)
    Generate(ctx context.Context, r *models.GenerateRequest) (*models.GenerateResponse, error)
    Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error)
    Embeddings(ctx context.Context, cr *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error)
}
```

**When adding new components:**
- Define the interface in an `iface.go` or `interfaces.go` file
- Keep interfaces focused and minimal
- Use `context.Context` as the first parameter

### 2. Optional Config Pattern

Constructors accept variadic optional configuration with nil-safe defaults.

**Example**: `llm/openai/client.go`
```go
func NewClient(optionalConfig ...*models.LLMConfig) (*Client, error) {
    var config *models.LLMConfig
    if len(optionalConfig) > 0 && optionalConfig[0] != nil {
        config = optionalConfig[0]
    } else {
        config = &models.LLMConfig{}
    }
    
    // Use config...
}
```

**Pattern benefits:**
- Optional configuration without multiple constructors
- Maintains backward compatibility
- Allows zero-value instantiation: `NewClient()`

### 3. Environment Variable Fallbacks

API keys and URLs fall back to environment variables when not explicitly configured.

**Example**: `llm/openai/client.go`
```go
apiKey := config.ApiKey
if apiKey == "" {
    apiKey = os.Getenv("OPENAI_API_KEY")
}

baseUrl := config.Url
if baseUrl == "" {
    baseUrl = os.Getenv("OPENAI_URL")
    if baseUrl == "" {
        baseUrl = OpenAI_API_URL_v1
    }
}
```

**Standard environment variable naming:**
- `{PROVIDER}_API_KEY` for API keys
- `{PROVIDER}_URL` or `{PROVIDER}_HOST` for endpoints
- Use uppercase with underscores

### 4. Context-First Parameters

All operations that perform I/O accept `context.Context` as the first parameter.

**Example**: `llm/iface/iface.go`
```go
Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error)
```

**When to use context:**
- Any network request
- Database operations
- Any operation that may need cancellation or timeout

### 5. Compile-Time Interface Checks

Use the blank identifier pattern to ensure types implement interfaces at compile time.

**Example**: `llm/openai/client.go`
```go
// Ensure Client implements iface.LLM
var _ iface.LLM = (*Client)(nil)
```

**Add this after the type definition:**
```go
type Client struct {
    client *openai.Client
}

// Compile-time interface check
var _ iface.LLM = (*Client)(nil)
```

### 6. Versioning Strategy

Use directory-based versioning for backward compatibility during major changes.

**Current versions:**
- `vectordb/v0/` - Legacy vector database interfaces
- `vectordb/v1/` - Current vector database with improved schema
- `rag/v1/` - Basic RAG implementation
- `rag/v2/` - Advanced RAG with callbacks and system-level abstractions

**When to version:**
- Breaking interface changes
- Major architectural shifts
- Keep old versions for backward compatibility
- Don't delete old versions unless absolutely necessary

### 7. Streaming Support Pattern

Optional streaming via variadic callback functions.

**Example**: `llm/iface/iface.go`
```go
Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error)
```

**Implementation pattern**:
```go
func (c *Client) Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error) {
    req := // build request
    
    // Check if streaming is requested
    if len(stream) > 0 && stream[0] != nil {
        return c.streamChat(ctx, req, stream[0])
    }
    
    // Non-streaming path
    resp, err := c.client.CreateChatCompletion(ctx, req)
    // ...
}
```

---

## Package Structure

```
ai-flow/
├── llm/                    # LLM client abstractions and implementations
│   ├── iface/              # Core LLM interface
│   ├── models/             # Shared request/response models
│   ├── openai/             # OpenAI implementation
│   ├── ollama/             # Ollama implementation
│   └── thinking/           # Extended thinking capabilities
│
├── rag/                    # Retrieval-Augmented Generation
│   ├── v1/                 # Basic RAG implementation
│   └── v2/                 # Advanced RAG with callbacks
│       ├── interfaces.go   # Retriever, Synthesizer, QueryEngine
│       ├── engine.go       # Query engine implementation
│       ├── retriever.go    # Vector retriever
│       ├── synthesizer.go  # Response synthesizer
│       ├── system.go       # High-level RAG system
│       └── reader/         # Document readers
│
├── vectordb/               # Vector database abstractions
│   ├── v0/                 # Legacy interfaces
│   │   ├── iface/          # Old interface definitions
│   │   ├── go-chromem/     # Old chromem wrapper
│   │   └── models/         # Old document models
│   └── v1/                 # Current implementation
│       ├── iface.go        # VectorStore interface
│       ├── schema/         # Node, Query, Response schemas
│       ├── simple.go       # Simple in-memory store
│       └── chromem/        # Chromem-go adapter
│
├── textsplitter/           # Text chunking utilities
│   ├── iface.go            # Tokenizer and splitter interfaces
│   ├── sentence_splitter.go    # Main splitter implementation
│   ├── paragraph_splitter.go   # Paragraph-level splitting
│   ├── tokenizer.go        # TikToken tokenizer
│   └── splitter_strategy.go    # Regex and Neurosnap strategies
│
├── http/                   # HTTP client utilities
│   ├── base_client.go      # Base HTTP client
│   ├── json_client.go      # JSON-specific client
│   └── http.go             # HTTP helpers
│
├── dotenv/                 # Environment variable loading
│   └── dotenv.go           # Auto-loads .env from project root
│
├── mocks/                  # Generated mocks
│   └── llm/                # LLM mocks
│
└── examples/               # Usage examples
    ├── dotenv/             # Dotenv usage
    ├── llm/                # LLM examples
    │   ├── openai/chat/
    │   └── ollama/chat/
    ├── rag/                # RAG examples
    │   ├── v1/
    │   └── v2/
    └── textsplitter/       # Text splitting examples
```

---

## Adding New LLM Providers

Use this checklist when adding a new LLM provider (e.g., Anthropic, Cohere, etc.).

### Step 1: Create Provider Package

Create a new directory: `llm/{provider}/`

### Step 2: Define Client Structure

```go
package provider

import (
    "context"
    "github.com/aqua777/ai-nexus/llm/iface"
    "github.com/aqua777/ai-nexus/llm/models"
)

type Client struct {
    config *models.LLMConfig
    // Add provider-specific client here
}

// Compile-time interface check
var _ iface.LLM = (*Client)(nil)
```

### Step 3: Implement Constructor

Follow the optional config pattern with environment variable fallbacks:

```go
const (
    PROVIDER_API_KEY_ENV = "PROVIDER_API_KEY"
    PROVIDER_URL_ENV     = "PROVIDER_URL"
    DEFAULT_PROVIDER_URL = "https://api.provider.com/v1"
)

func NewClient(optionalConfig ...*models.LLMConfig) (*Client, error) {
    var config *models.LLMConfig
    if len(optionalConfig) > 0 && optionalConfig[0] != nil {
        config = optionalConfig[0]
    } else {
        config = &models.LLMConfig{}
    }

    // Get API key with fallback
    apiKey := config.ApiKey
    if apiKey == "" {
        apiKey = os.Getenv(PROVIDER_API_KEY_ENV)
    }
    if apiKey == "" {
        return nil, errors.New("API key required")
    }

    // Get URL with fallback
    baseUrl := config.Url
    if baseUrl == "" {
        baseUrl = os.Getenv(PROVIDER_URL_ENV)
        if baseUrl == "" {
            baseUrl = DEFAULT_PROVIDER_URL
        }
    }

    // Initialize provider client
    client := // provider-specific initialization

    return &Client{
        config: config,
        client: client,
    }, nil
}
```

### Step 4: Implement Interface Methods

Implement all four methods from `llm/iface/iface.go`:

#### ListModels
```go
func (c *Client) ListModels(ctx context.Context) ([]*models.Model, error) {
    // Call provider API
    resp, err := c.client.ListModels(ctx)
    if err != nil {
        return nil, err
    }

    // Map to ai-flow models
    var result []*models.Model
    for _, m := range resp.Models {
        result = append(result, &models.Model{
            ID:    m.ID,
            Name:  m.Name,
            Model: m.Model,
        })
    }
    return result, nil
}
```

#### Generate
```go
func (c *Client) Generate(ctx context.Context, r *models.GenerateRequest) (*models.GenerateResponse, error) {
    // Map ai-flow request to provider format
    req := // provider-specific request
    
    resp, err := c.client.Generate(ctx, req)
    if err != nil {
        return nil, err
    }

    // Map provider response to ai-flow format
    return &models.GenerateResponse{
        Text:             resp.Text,
        Model:            resp.Model,
        PromptTokens:     resp.Usage.PromptTokens,
        CompletionTokens: resp.Usage.CompletionTokens,
        TotalTokens:      resp.Usage.TotalTokens,
    }, nil
}
```

#### Chat (with streaming support)
```go
func (c *Client) Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error) {
    // Map messages
    providerMessages := make([]ProviderMessage, len(r.Messages))
    for i, msg := range r.Messages {
        providerMessages[i] = ProviderMessage{
            Role:    string(msg.Role),
            Content: msg.Content,
        }
    }

    req := ProviderChatRequest{
        Model:    r.Model,
        Messages: providerMessages,
    }

    // Handle streaming
    if len(stream) > 0 && stream[0] != nil {
        return c.streamChat(ctx, req, stream[0])
    }

    // Non-streaming path
    resp, err := c.client.CreateChat(ctx, req)
    if err != nil {
        return nil, err
    }

    return &models.ChatResponse{
        Content: resp.Content,
        Metadata: &models.ChatResponseMetadata{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
    }, nil
}

func (c *Client) streamChat(ctx context.Context, req ProviderChatRequest, callback func(chunk []byte) error) (*models.ChatResponse, error) {
    stream, err := c.client.CreateChatStream(ctx, req)
    if err != nil {
        return nil, err
    }
    defer stream.Close()

    var fullContent string
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        if chunk.Delta != "" {
            fullContent += chunk.Delta
            if err := callback([]byte(chunk.Delta)); err != nil {
                return nil, err
            }
        }
    }

    return &models.ChatResponse{
        Content: fullContent,
    }, nil
}
```

#### Embeddings
```go
func (c *Client) Embeddings(ctx context.Context, cr *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
    model := cr.Model
    if model == "" {
        model = "default-embedding-model"
    }

    resp, err := c.client.CreateEmbedding(ctx, ProviderEmbeddingRequest{
        Input: cr.Content,
        Model: model,
    })
    if err != nil {
        return nil, err
    }

    return &models.EmbeddingsResponse{
        Embeddings: resp.Embedding,
    }, nil
}
```

### Step 5: Add Example

Create `examples/llm/{provider}/chat/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    _ "github.com/aqua777/ai-nexus/dotenv"
    "github.com/aqua777/ai-nexus/llm/models"
    "github.com/aqua777/ai-nexus/llm/{provider}"
)

func main() {
    ctx := context.Background()
    
    client, err := provider.NewClient()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    req := &models.ChatRequest{
        Model: "model-name",
        Messages: []*models.Message{
            {Role: models.SystemRole, Content: "You are a helpful assistant."},
            {Role: models.UserRole, Content: "Hello!"},
        },
    }

    resp, err := client.Chat(ctx, req)
    if err != nil {
        log.Fatalf("Failed to chat: %v", err)
    }

    fmt.Println("Response:", resp.Content)
}
```

### Step 6: Add Tests

See [Testing Conventions](#testing-conventions) section below.

---

## Testing Conventions

This project uses `testify/suite` for all tests. Follow these patterns strictly.

### Test File Structure

```go
package mypackage

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

// 1. Define test suite struct
type MyPackageTestSuite struct {
    suite.Suite
    // Add test fixtures here
    client *Client
    tempDir string
}

// 2. Test runner function
func TestMyPackageTestSuite(t *testing.T) {
    suite.Run(t, new(MyPackageTestSuite))
}

// 3. Setup methods (optional)
func (s *MyPackageTestSuite) SetupSuite() {
    // Runs once before all tests
    s.tempDir = s.T().TempDir()
}

func (s *MyPackageTestSuite) SetupTest() {
    // Runs before each test
    s.client = NewClient()
}

// 4. Test methods
func (s *MyPackageTestSuite) TestFeature_HappyPath() {
    result, err := s.client.DoSomething()
    
    s.NoError(err)
    s.NotNil(result)
    s.Equal("expected", result.Value)
}

func (s *MyPackageTestSuite) TestFeature_ErrorCase() {
    result, err := s.client.DoSomethingInvalid()
    
    s.Error(err)
    s.Nil(result)
    s.Contains(err.Error(), "expected error message")
}
```

### Mock Implementations

Create mocks inline in test files for testing interfaces:

**Example**: `rag/v2/engine_test.go`
```go
// Mock Retriever
type MockRetriever struct {
    Nodes []schema.NodeWithScore
    Err   error
}

func (m *MockRetriever) Retrieve(ctx context.Context, query schema.QueryBundle) ([]schema.NodeWithScore, error) {
    return m.Nodes, m.Err
}

// Mock LLM
type MockLLM struct {
    ChatResponse string
    Embedding    []float32
}

var _ iface.LLM = (*MockLLM)(nil)

func (m *MockLLM) Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error) {
    return &models.ChatResponse{Content: m.ChatResponse}, nil
}

func (m *MockLLM) Embeddings(ctx context.Context, cr *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
    return &models.EmbeddingsResponse{Embeddings: m.Embedding}, nil
}

// Implement other interface methods...
```

### Test Naming Conventions

- Test suite struct: `{Package}TestSuite`
- Test runner: `Test{Package}TestSuite(t *testing.T)`
- Test methods: `Test{Feature}_{Scenario}`

**Examples:**
- `TestClient_Chat_Success`
- `TestRetriever_Retrieve_EmptyQuery`
- `TestVectorStore_Query_InvalidEmbedding`

### Assertion Patterns

Use testify assertions (not `t.Error`):

```go
// Good
s.NoError(err)
s.Nil(result)
s.Equal(expected, actual)
s.Contains(str, substr)
s.Len(slice, 5)

// Bad (don't use these)
if err != nil {
    t.Errorf("unexpected error: %v", err)
}
```

### Testing Happy Path AND Error Cases

Always test both:

```go
func (s *MyTestSuite) TestFeature_Success() {
    result, err := s.client.Feature(validInput)
    s.NoError(err)
    s.NotNil(result)
}

func (s *MyTestSuite) TestFeature_InvalidInput() {
    result, err := s.client.Feature(invalidInput)
    s.Error(err)
    s.Contains(err.Error(), "invalid input")
}

func (s *MyTestSuite) TestFeature_NetworkError() {
    // Setup mock to return error
    result, err := s.client.Feature(input)
    s.Error(err)
}
```

### Full Integration Tests

Include full end-to-end tests that exercise the entire flow:

**Example**: `rag/v2/engine_test.go`
```go
func (s *EngineTestSuite) TestFullRAGFlow() {
    ctx := context.Background()

    // 1. Setup Components
    mockLLM := &MockLLM{
        Embedding:    []float32{0.1, 0.2, 0.3},
        ChatResponse: "This is a generated answer.",
    }
    vectorStore := store.NewSimpleVectorStore()

    // 2. Add Documents
    nodes := []schema.Node{
        {ID: "1", Text: "Document content...", Embedding: []float64{0.1, 0.2, 0.3}},
    }
    _, err := vectorStore.Add(ctx, nodes)
    s.NoError(err)

    // 3. Create Components
    retriever := NewVectorRetriever(vectorStore, mockLLM, "model", 1)
    synthesizer := NewSimpleSynthesizer(mockLLM, "model")
    engine := NewRetrieverQueryEngine(retriever, synthesizer)

    // 4. Execute Query
    query := schema.QueryBundle{QueryString: "test query"}
    response, err := engine.Query(ctx, query)

    // 5. Verify
    s.NoError(err)
    s.Equal("This is a generated answer.", response.Response)
    s.Len(response.SourceNodes, 1)
}
```

---

## Example Patterns

Examples live in `examples/` and demonstrate real-world usage.

### Example Structure

```
examples/
├── go.mod              # Separate go.mod for examples
└── {feature}/
    └── {scenario}/
        └── main.go     # Runnable example
```

### Standard Example Template

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    _ "github.com/aqua777/ai-nexus/dotenv"  // Auto-load .env
    "github.com/aqua777/ai-nexus/llm/models"
    "github.com/aqua777/ai-nexus/llm/openai"
)

func main() {
    ctx := context.Background()
    
    // 1. Create client
    client, err := openai.NewClient()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // 2. Build request
    req := &models.ChatRequest{
        Model: "gpt-4o-mini",
        Messages: []*models.Message{
            {Role: models.SystemRole, Content: "You are helpful."},
            {Role: models.UserRole, Content: "Hello!"},
        },
    }

    // 3. Execute
    resp, err := client.Chat(ctx, req)
    if err != nil {
        log.Fatalf("Failed to chat: %v", err)
    }

    // 4. Use response
    fmt.Println("Response:", resp.Content)
}
```

### Dotenv Import Pattern

Always import dotenv with blank identifier at the top:

```go
import (
    "context"
    "fmt"
    
    _ "github.com/aqua777/ai-nexus/dotenv"  // ← Blank import
    "github.com/aqua777/ai-nexus/llm/openai"
)
```

This auto-loads `.env` from the project root via `init()` function.

### Example Best Practices

1. **Keep it minimal**: Focus on one feature
2. **Clear error handling**: Use `log.Fatalf` for unrecoverable errors
3. **Use context.Background()**: Examples don't need complex context
4. **Print results**: Show the output so users can see it works
5. **Comment non-obvious parts**: Help users understand what's happening

---

## Configuration Guidelines

### LLMConfig Structure

Defined in `llm/models/llm_config.go`:

```go
type LLMConfig struct {
    Url    string `json:"url"`
    ApiKey string `json:"api_key"`
}
```

### Configuration Priority Order

1. **Explicit config** passed to constructor
2. **Environment variables** as fallback
3. **Default values** as final fallback

**Example implementation**:
```go
func NewClient(optionalConfig ...*models.LLMConfig) (*Client, error) {
    var config *models.LLMConfig
    if len(optionalConfig) > 0 && optionalConfig[0] != nil {
        config = optionalConfig[0]  // 1. Explicit
    } else {
        config = &models.LLMConfig{}
    }

    apiKey := config.ApiKey
    if apiKey == "" {
        apiKey = os.Getenv("API_KEY")  // 2. Environment
    }
    if apiKey == "" {
        return nil, errors.New("API key required")  // 3. No default for required
    }

    baseUrl := config.Url
    if baseUrl == "" {
        baseUrl = os.Getenv("API_URL")  // 2. Environment
        if baseUrl == "" {
            baseUrl = DEFAULT_URL  // 3. Default
        }
    }
    
    // Use apiKey and baseUrl...
}
```

### Configuration for Complex Systems

For complex systems like RAG, define dedicated config structs:

**Example**: `rag/v2/system.go`
```go
type RAGConfig struct {
    OpenAIKey      string
    OpenAIBaseURL  string
    LLMModel       string
    EmbeddingModel string
    ChunkSize      int
    ChunkOverlap   int
    TopK           int
    PersistPath    string
    CollectionName string
}

// Provide defaults
func DefaultRAGConfig() *RAGConfig {
    return &RAGConfig{
        LLMModel:       "gpt-4o-mini",
        EmbeddingModel: "text-embedding-3-small",
        ChunkSize:      1024,
        ChunkOverlap:   200,
        TopK:           5,
        CollectionName: "default-rag",
    }
}
```

---

## Code Organization Best Practices

### 1. Interface Separation

Keep interfaces separate from implementations:

```
package/
├── iface.go        # Interface definitions
├── impl.go         # Implementation
└── models.go       # Data structures
```

Or for larger packages:

```
package/
├── iface/
│   └── iface.go    # Interfaces
├── models/
│   └── models.go   # Shared types
└── impl/
    └── impl.go     # Implementation
```

### 2. Models Package Pattern

Shared types go in `models/` subdirectory:

```
llm/
├── models/
│   ├── chat.go           # Chat-related types
│   ├── common.go         # Common types (Role, etc.)
│   ├── embeddings.go     # Embedding types
│   ├── generate.go       # Generate types
│   ├── llm_config.go     # Config types
│   ├── model.go          # Model metadata
│   └── options.go        # Request options
└── openai/
    └── client.go         # Uses types from models/
```

### 3. File Naming Conventions

- `{feature}.go` - Main implementation
- `{feature}_test.go` - Tests
- `iface.go` or `interfaces.go` - Interface definitions
- `models.go` - Data structures
- `utils.go` - Helper functions
- `{feature}_strategy.go` - Strategy pattern implementations

### 4. Package Naming

- Use **singular** package names: `llm` not `llms`
- Use **descriptive** names: `textsplitter` not `splitter`
- Avoid **stutter**: Don't export `llm.LLMClient`, prefer `llm.Client`

### 5. Versioning Changes

When making breaking changes:

```
package/
├── v0/          # Old version (keep for compatibility)
│   └── ...
├── v1/          # Current stable
│   └── ...
└── v2/          # New development
    └── ...
```

- Never delete old versions
- Update documentation to point to latest
- Provide migration guides in comments or README

### 6. Import Grouping

Standard Go import order:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "os"

    // 2. Third-party packages
    "github.com/google/uuid"
    "github.com/stretchr/testify/suite"

    // 3. Local packages
    "github.com/aqua777/ai-nexus/llm/iface"
    "github.com/aqua777/ai-nexus/llm/models"
)
```

### 7. Error Handling

- Return errors, don't panic
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Use descriptive error messages
- Define sentinel errors for expected conditions

```go
var (
    ErrAPIKeyRequired = errors.New("API key is required")
    ErrInvalidModel   = errors.New("invalid model specified")
)

func (c *Client) Chat(ctx context.Context, r *ChatRequest) (*ChatResponse, error) {
    if r.Model == "" {
        return nil, ErrInvalidModel
    }
    
    resp, err := c.client.Chat(ctx, r)
    if err != nil {
        return nil, fmt.Errorf("chat request failed: %w", err)
    }
    
    return resp, nil
}
```

---

## Quick Reference Checklist

When adding new features, ensure you:

- [ ] Define interfaces before implementations
- [ ] Use optional config pattern for constructors
- [ ] Add environment variable fallbacks
- [ ] Include `context.Context` as first parameter
- [ ] Add compile-time interface checks
- [ ] Write tests using testify/suite
- [ ] Create both happy path and error tests
- [ ] Add examples in `examples/` directory
- [ ] Follow error handling conventions
- [ ] Use proper import grouping
- [ ] Document exported functions and types
- [ ] Version breaking changes appropriately

---

## Summary

This codebase values:
- **Consistency** over cleverness
- **Interfaces** over concrete types
- **Explicit** over implicit configuration
- **Testing** as a first-class citizen
- **Examples** that actually run

When in doubt, look at existing implementations (particularly `llm/openai/client.go` and `rag/v2/`) as reference patterns.

Happy coding! 🚀

