# AI Nexus
AI Nexus is a Go toolkit for building LLM-powered applications. It provides:

- LLM provider abstraction (OpenAI, Ollama)
- A production-ready text splitter with recursive strategies
- A simple vector database interface with a Chromem-backed store
- A modular RAG system (v2) for ingestion and querying
- Lightweight HTTP helpers for JSON and streaming
- Automatic .env loader for local development

Keep this README high-level; see subpackage READMEs and examples for details.

## Features

- **LLM abstraction**: A single `LLM` interface with implementations for **OpenAI** and **Ollama**.
- **Text splitting**: Recursive, tokenizer-aware splitter with regex and Neurosnap sentence strategies. See the dedicated README.
- **Vector DB**: Simple pluggable store API with a **Chromem** implementation (in-memory or persistent).
- **RAG v2**: Composable ingestion and query pipeline with callbacks and configurable chunking/top-k.
- **HTTP utilities**: Typed status/method constants, JSON helpers, and newline-delimited JSON streaming support.
- **dotenv**: Auto-load `.env` from the current or parent directories (up to the project root).

## Repository structure

- **dotenv/** – Auto-loading of environment variables at startup
- **http/** – Minimal HTTP client/helpers for JSON and streaming
- **llm/** – Provider-agnostic interfaces and models
  - **openai/** – OpenAI client implementation
  - **ollama/** – Ollama client implementation
  - **models/** – Shared request/response models and config
  - **iface/** – `LLM` interface definition
  - **thinking/** – Thought helpers/utilities
- **rag/** – Retrieval Augmented Generation
  - **v2/** – RAG system with ingestion, retrieval, and synthesis
  - v1 – Older iteration maintained for reference
- **textsplitter/** – Advanced text splitting package
  - See its README for usage and configuration
- **vectordb/** – Vector store interfaces and implementations
  - **v1/chromem/** – Chromem-backed store (in-memory or persistent)
  - v0 – Early API kept for reference
- **examples/** – End-to-end runnable samples

## Installation and requirements

- Requires Go 1.24+
- Import the packages you need; Go will resolve the module automatically. For example:

```go
import (
    "github.com/aqua777/ai-nexus/llm/openai"
    "github.com/aqua777/ai-nexus/llm/models"
)
```

## Configuration

- **Environment variables** (honored via `llm/models.LLMConfig`):
  - `OPENAI_API_KEY`, `OPENAI_URL` (defaults to https://api.openai.com/v1)
  - `OLLAMA_URL` (defaults to http://localhost:11434)
- **dotenv**: If you import `github.com/aqua777/ai-nexus/dotenv`, `.env` is auto-loaded at init.

## Quick links

- **Text Splitter**
  - README: [textsplitter/README.md](./textsplitter/README.md)
  - Example: [examples/textsplitter/sentence-splitter](./examples/textsplitter/sentence-splitter)

- **LLM (OpenAI)**
  - Package: [llm/openai](./llm/openai)
  - Example (chat): [examples/llm/openai/chat](./examples/llm/openai/chat)

- **LLM (Ollama)**
  - Package: [llm/ollama](./llm/ollama)
  - Examples:
    - Chat: [examples/llm/ollama/chat](./examples/llm/ollama/chat)
    - Generate: [examples/llm/ollama/generate](./examples/llm/ollama/generate)
    - List models: [examples/llm/ollama/list_models](./examples/llm/ollama/list_models)

- **RAG**
  - v1: [rag/v1](./rag/v1), examples: [examples/rag/v1](./examples/rag/v1)
  - v2 System: [rag/v2](./rag/v2), examples: [examples/rag/v2](./examples/rag/v2)

- **Vector DB**
  - Chroma-like embedable in-memory or persistent vector database: [vectordb/v1/chromem](./vectordb/v1/chromem)

- **HTTP**
  - Package: [http](./http)

- **dotenv**
  - Package: [dotenv](./dotenv)
  - Example: [examples/dotenv](./examples/dotenv)

## Example: OpenAI chat (see full example for details)

Refer to [examples/llm/openai/chat](./examples/llm/openai/chat) for a runnable program. High-level flow:

1. Import `dotenv` to auto-load `.env` (optional but convenient).
2. Create an OpenAI client and `ChatRequest`.
3. Call `Chat` and use the response content.

## Acknowledgements

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)
- [pkoukk/tiktoken-go](https://github.com/pkoukk/tiktoken-go)
- [neurosnap/sentences](https://github.com/neurosnap/sentences)
- [philippgille/chromem-go](https://github.com/philippgille/chromem-go)
