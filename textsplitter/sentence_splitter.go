package textsplitter

import (
	"fmt"
	"strings"
)

const (
	DefaultChunkSize     = 1024
	DefaultChunkOverlap  = 200
	DefaultParagraphSep  = "\n\n\n"
	DefaultSeparator     = " "
	DefaultChunkingRegex = `[^,.;。？！]+[,.;。？！]?|[,.;。？！]`
)

// textSplit holds intermediate split information.
type textSplit struct {
	text       string
	isSentence bool
	tokenSize  int
}

// SentenceSplitter splits text with a preference for complete sentences.
type SentenceSplitter struct {
	ChunkSize              int
	ChunkOverlap           int
	Separator              string
	ParagraphSeparator     string
	SecondaryChunkingRegex string
	Tokenizer              Tokenizer
	SplitterStrategy       SentenceSplitterStrategy

	_splitFns            []func(string) []string
	_subSentenceSplitFns []func(string) []string
}

// NewSentenceSplitter creates a new SentenceSplitter.
// Pass 0 or empty strings to use defaults.
// If tokenizer is nil, defaults to SimpleTokenizer.
// If splitterStrategy is nil, defaults to RegexSplitterStrategy with DefaultChunkingRegex.
func NewSentenceSplitter(
	chunkSize int,
	chunkOverlap int,
	tokenizer Tokenizer,
	splitterStrategy SentenceSplitterStrategy,
) *SentenceSplitter {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	// chunkOverlap can be 0. We do not default it if 0 is passed.
	// To use default overlap, caller should pass DefaultChunkOverlap.

	if tokenizer == nil {
		tokenizer = NewSimpleTokenizer()
	}

	if splitterStrategy == nil {
		splitterStrategy = NewRegexSplitterStrategy(DefaultChunkingRegex)
	}

	s := &SentenceSplitter{
		ChunkSize:              chunkSize,
		ChunkOverlap:           chunkOverlap,
		Separator:              DefaultSeparator,
		ParagraphSeparator:     DefaultParagraphSep,
		SecondaryChunkingRegex: DefaultChunkingRegex,
		Tokenizer:              tokenizer,
		SplitterStrategy:       splitterStrategy,
	}

	s.initSplitFns()
	return s
}

func (s *SentenceSplitter) initSplitFns() {
	// Primary split functions:
	// 1. Paragraph separator
	// 2. Sentence Splitter Strategy (Regex or Neurosnap or custom)
	s._splitFns = []func(string) []string{
		SplitBySep(s.ParagraphSeparator),
		func(text string) []string { return s.SplitterStrategy.Split(text) },
	}

	// Sub-sentence split functions (fallback if sentences are still too big):
	// 1. Regex fallback (hardcoded default regex often used as backup)
	// 2. Separator (Word)
	// 3. Character
	// Note: Python implementation allows customizing secondary regex.
	s._subSentenceSplitFns = []func(string) []string{
		SplitByRegex(s.SecondaryChunkingRegex),
		SplitBySep(s.Separator),
		SplitByChar(),
	}
}

// SplitText splits the text into chunks.
func (s *SentenceSplitter) SplitText(text string) []string {
	return s.splitText(text, s.ChunkSize)
}

func (s *SentenceSplitter) splitText(text string, chunkSize int) []string {
	if text == "" {
		return []string{text}
	}

	splits := s.split(text, chunkSize)
	chunks := s.merge(splits, chunkSize)
	return s.postprocessChunks(chunks)
}

func (s *SentenceSplitter) split(text string, chunkSize int) []textSplit {
	tokenSize := s.getTokenSize(text)
	if tokenSize <= chunkSize {
		return []textSplit{{text: text, isSentence: true, tokenSize: tokenSize}}
	}

	textSplitsByFns, isSentence := s.getSplitsByFns(text)
	var textSplits []textSplit

	for _, splitStr := range textSplitsByFns {
		tokenSize := s.getTokenSize(splitStr)
		if tokenSize <= chunkSize {
			textSplits = append(textSplits, textSplit{
				text:       splitStr,
				isSentence: isSentence,
				tokenSize:  tokenSize,
			})
		} else {
			recursiveSplits := s.split(splitStr, chunkSize)
			textSplits = append(textSplits, recursiveSplits...)
		}
	}
	return textSplits
}

func (s *SentenceSplitter) merge(splits []textSplit, chunkSize int) []string {
	var chunks []string
	// current chunk buffer: list of (text, length)
	type bufItem struct {
		text string
		len  int
	}
	var curChunk []bufItem
	var lastChunk []bufItem
	curChunkLen := 0
	newChunk := true

	closeChunk := func() {
		var sb strings.Builder
		for _, item := range curChunk {
			sb.WriteString(item.text)
		}
		chunks = append(chunks, sb.String())

		lastChunk = curChunk
		curChunk = nil // reset
		curChunkLen = 0
		newChunk = true

		// Add overlap from lastChunk
		if len(lastChunk) > 0 {
			lastIndex := len(lastChunk) - 1
			for lastIndex >= 0 {
				item := lastChunk[lastIndex]
				if curChunkLen+item.len <= s.ChunkOverlap {
					curChunkLen += item.len
					// Prepend to curChunk
					curChunk = append([]bufItem{item}, curChunk...)
					lastIndex--
				} else {
					break
				}
			}
		}
	}

	splitIdx := 0
	for splitIdx < len(splits) {
		curSplit := splits[splitIdx]
		if curSplit.tokenSize > chunkSize {
			// Should not happen if recursion worked, but safety check
			// In Python it raises ValueError.
			// We'll just force it in or panic.
			// For now, panic to be noticeable during dev.
			panic(fmt.Sprintf("Single token exceeded chunk size: %d > %d", curSplit.tokenSize, chunkSize))
		}

		if curChunkLen+curSplit.tokenSize > chunkSize && !newChunk {
			closeChunk()
		} else {
			// If new chunk with overlap and adding split exceeds chunk size, remove overlap
			if newChunk && curChunkLen+curSplit.tokenSize > chunkSize {
				for len(curChunk) > 0 && curChunkLen+curSplit.tokenSize > chunkSize {
					// Pop from front
					removed := curChunk[0]
					curChunk = curChunk[1:]
					curChunkLen -= removed.len
				}
			}

			// Add split if it fits or if it is a sentence or if it's a new chunk (must add at least one)
			if curSplit.isSentence || curChunkLen+curSplit.tokenSize <= chunkSize || newChunk {
				curChunkLen += curSplit.tokenSize
				curChunk = append(curChunk, bufItem{text: curSplit.text, len: curSplit.tokenSize})
				splitIdx++
				newChunk = false
			} else {
				closeChunk()
			}
		}
	}

	if !newChunk {
		var sb strings.Builder
		for _, item := range curChunk {
			sb.WriteString(item.text)
		}
		chunks = append(chunks, sb.String())
	}

	return chunks
}

func (s *SentenceSplitter) postprocessChunks(chunks []string) []string {
	var newChunks []string
	for _, chunk := range chunks {
		stripped := strings.TrimSpace(chunk)
		if stripped == "" {
			continue
		}
		newChunks = append(newChunks, stripped)
	}
	return newChunks
}

func (s *SentenceSplitter) getTokenSize(text string) int {
	return len(s.Tokenizer.Encode(text))
}

func (s *SentenceSplitter) getSplitsByFns(text string) ([]string, bool) {
	for _, splitFn := range s._splitFns {
		splits := splitFn(text)
		if len(splits) > 1 {
			return splits, true
		}
	}

	var splits []string
	for _, splitFn := range s._subSentenceSplitFns {
		splits = splitFn(text)
		if len(splits) > 1 {
			break
		}
	}
	return splits, false
}
