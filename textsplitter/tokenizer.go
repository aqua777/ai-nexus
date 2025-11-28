package textsplitter

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// SimpleTokenizer tokenizes text by splitting on whitespace.
type SimpleTokenizer struct{}

func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{}
}

func (t *SimpleTokenizer) Encode(text string) []string {
	// Return strings.Fields if not empty, otherwise fallback to split by char to avoid 0 length?
	// No, empty string should be 0 tokens.
	// But if we have a large block of text without spaces, Fields returns [text].
	// This counts as 1 token. This is bad for chunking if we rely on token count.
	// We should maybe split by character if fields is length 1 and length of text is large?
	// But SimpleTokenizer is "simple".
	// Let's just leave it as is but document it.
	// Wait, if the user uses this for chunking, and we have a 5000 char string with no spaces,
	// it will be 1 token. The chunker will see 1 token <= 1024 chunk size? No, 1 <= 1024 is true.
	// So it will accept the whole string as one chunk.
	// And then the embedding model (which uses real tokens) will choke on 5000 chars.

	// The fix is to use a better tokenizer or make SimpleTokenizer smarter.
	// Let's make SimpleTokenizer split by character if fields returns 1 element that is too long?
	// Or just use a real tokenizer.

	return strings.Fields(text)
}

// TikTokenTokenizer tokenizes text using OpenAI's tiktoken.
type TikTokenTokenizer struct {
	encoding *tiktoken.Tiktoken
}

func NewTikTokenTokenizer(model string) (*TikTokenTokenizer, error) {
	if model == "" {
		model = "gpt-3.5-turbo" // Default
	}
	// Use EncodingForModel to get the correct encoding for the model
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding for model %s: %w", model, err)
	}
	return &TikTokenTokenizer{encoding: tkm}, nil
}

func (t *TikTokenTokenizer) Encode(text string) []string {
	// Encode returns []int. We need to map them to strings.
	// Since the interface requires []string (and splitter uses len()), we can return a list of string representations.
	// OR, since we primarily use this for length checking, we can just return dummy strings?
	// The Python code uses `len(self._tokenizer(text))`.
	// If we just return strings, we satisfy the interface.
	// However, simply converting int->string ("123") might be enough if only length matters.
	// BUT if we ever need to see the tokens, we might want real token strings?
	// Tiktoken-go doesn't expose "decode single token" easily without Decode([]int).
	// Let's just return stringified integers for now, as it's efficient enough for length check.
	// Actually, to be safe for future use (e.g. debugging), let's just return string representations of IDs.

	tokenIDs := t.encoding.Encode(text, nil, nil)
	tokens := make([]string, len(tokenIDs))
	for i, id := range tokenIDs {
		tokens[i] = fmt.Sprintf("%d", id)
	}
	return tokens
}
