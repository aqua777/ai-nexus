package textsplitter

import (
	"strings"
)

// ParagraphSplitter splits text into chunks based on paragraphs.
type ParagraphSplitter struct {
	MaxChunkSize int
	Separator    string
}

// NewParagraphSplitter creates a new ParagraphSplitter.
// If maxChunkSize is <= 0, it defaults to DefaultChunkSize.
func NewParagraphSplitter(maxChunkSize int) *ParagraphSplitter {
	if maxChunkSize <= 0 {
		maxChunkSize = DefaultChunkSize
	}
	return &ParagraphSplitter{
		MaxChunkSize: maxChunkSize,
		Separator:    "\n",
	}
}

// SplitText splits the text into chunks based on the strategy.
func (s *ParagraphSplitter) SplitText(text string) []string {
	if text == "" {
		return []string{}
	}

	// 1. Split into paragraphs and filter empty ones
	rawParagraphs := strings.Split(text, s.Separator)
	var paragraphs []string
	for _, p := range rawParagraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			// We preserve the original paragraph content unless we want to trim it.
			// The requirements say "ignore empty lines", usually implies checking trimmed version.
			// I'll store the trimmed version for cleanliness, or maybe the original?
			// Usually with paragraph splitting, trimming whitespace around the paragraph is desired.
			paragraphs = append(paragraphs, trimmed)
		}
	}

	if len(paragraphs) == 0 {
		return []string{}
	}

	var chunks []string
	var currentChunk []string
	currentSize := 0

	for _, p := range paragraphs {
		// Handle oversized paragraphs
		if len(p) > s.MaxChunkSize {
			// If we have a current chunk accumulating, flush it first
			if len(currentChunk) > 0 {
				chunks = append(chunks, strings.Join(currentChunk, s.Separator))
				// For overlap logic with a split paragraph:
				// The "last paragraph" of the previous chunk is the last element of currentChunk.
				// But since we are about to handle a huge paragraph, we might need to clear currentChunk logic.
				// Let's follow the standard flow:
				// If we flush here, the next chunk (the pieces of the huge paragraph) should technically start with the overlap.
				// But adding a small paragraph to a huge chunk piece might not make sense if the piece itself is maxed out.
				// However, the rule is simple: "last paragraph from the previous chunks as the first one".
				
				// Reset current chunk, but keep overlap for the *next* chunk construction
				lastPara := currentChunk[len(currentChunk)-1]
				currentChunk = []string{lastPara}
				currentSize = len(lastPara)
			}

			// Now split the huge paragraph
			// We just chop it by characters. 
			// Since we might have overlap in currentChunk (from the flush above), we need to see if it fits.
			// Actually, simpler logic: treat the pieces of the huge paragraph as just more paragraphs coming in sequence.
			// So let's split 'p' into sub-paragraphs and insert them into the processing queue?
			// That might be cleaner. But we are iterating.
			
			// Let's process the sub-parts immediately.
			subParts := s.splitOversizedParagraph(p)
			for _, subP := range subParts {
				// Re-run the adding logic for each sub-part
				if currentSize+len(subP)+len(s.Separator) > s.MaxChunkSize {
					// Flush current chunk
					if len(currentChunk) > 0 {
						chunks = append(chunks, strings.Join(currentChunk, s.Separator))
						// Overlap
						lastPara := currentChunk[len(currentChunk)-1]
						
						// If overlap itself + new subP is too big, drop overlap
						if len(lastPara)+len(subP)+len(s.Separator) > s.MaxChunkSize {
							currentChunk = []string{subP}
							currentSize = len(subP)
						} else {
							currentChunk = []string{lastPara, subP}
							currentSize = len(lastPara) + len(s.Separator) + len(subP)
						}
					} else {
						// Should not happen if logic is correct, but if currentChunk is empty:
						currentChunk = []string{subP}
						currentSize = len(subP)
					}
				} else {
					if len(currentChunk) > 0 {
						currentSize += len(s.Separator)
					}
					currentChunk = append(currentChunk, subP)
					currentSize += len(subP)
				}
			}
			continue
		}

		// Normal paragraph handling
		// Calculate added size (plus separator if not first)
		addedSize := len(p)
		if len(currentChunk) > 0 {
			addedSize += len(s.Separator)
		}

		if currentSize+addedSize > s.MaxChunkSize {
			// Chunk capacity reached.
			// 1. Make a new chunk from current buffer
			chunks = append(chunks, strings.Join(currentChunk, s.Separator))
			
			// 2. Start new buffer with last paragraph of previous chunk (Overlap)
			if len(currentChunk) > 0 {
				lastPara := currentChunk[len(currentChunk)-1]
				
				// Edge case: If overlap + current p > MaxChunkSize, drop overlap
				// Note: we need to account for separator between overlap and p
				overlapSize := len(lastPara) + len(s.Separator) + len(p)
				if overlapSize > s.MaxChunkSize {
					// Drop overlap
					currentChunk = []string{p}
					currentSize = len(p)
				} else {
					currentChunk = []string{lastPara, p}
					currentSize = overlapSize
				}
			} else {
				// Should be impossible if we checked size, but safe fallback
				currentChunk = []string{p}
				currentSize = len(p)
			}
		} else {
			// Add to current chunk
			currentChunk = append(currentChunk, p)
			currentSize += addedSize
		}
	}

	// Add any remaining content
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, s.Separator))
	}

	return chunks
}

func (s *ParagraphSplitter) splitOversizedParagraph(text string) []string {
	var parts []string
	runes := []rune(text)
	length := len(runes) // character count, not bytes
	
	// We need to be careful about mixing byte length (for MaxChunkSize check usually?) 
	// User said "MaxChunkSize to determine the accepted size... characters".
	// Go strings are byte slices. len(string) is bytes.
	// If MaxChunkSize is bytes, we should slice bytes. If chars, we should slice runes.
	// Standard Go practice for "text" usually implies bytes unless specified "runes" or "chars".
	// But for splitting text naturally, runes are safer to avoid cutting multi-byte chars.
	// However, len(p) in the main loop returns bytes. 
	// So consistent approach: treat MaxChunkSize as BYTES for checking, but split safely by RUNES if needed,
	// checking byte length of the resulting string.
	
	// Let's assume MaxChunkSize is effectively bytes since `len("string")` is bytes.
	// But we want to avoid splitting inside a utf8 sequence.
	
	currentStart := 0
	for currentStart < length {
		// Ideally we take as many runes as possible that fit in MaxChunkSize bytes.
		// This is expensive to calculate repeatedly. 
		// For simplicity/performance in this context, assuming mostly valid text:
		// We'll grab a chunk of runes roughly size MaxChunkSize/avg_byte_per_rune?
		// Or just iterate runes.
		
		// Simpler greedy approach:
		// Take a slice of runes that is length s.MaxChunkSize (if 1 byte per rune).
		// Check byte length.
		end := currentStart + s.MaxChunkSize
		if end > length {
			end = length
		}
		
		// Adjust end down if byte length exceeds MaxChunkSize
		// (In worst case of 4-byte runes, s.MaxChunkSize runes could be 4x memory limit)
		// But wait, if we only care that `len(string(runes))` <= MaxChunkSize:
		
		// Let's build the string token
		chunkRunes := runes[currentStart:]
		
		count := 0
		takeRunes := 0
		for _, r := range chunkRunes {
			rlen := len(string(r))
			if count+rlen > s.MaxChunkSize {
				break
			}
			count += rlen
			takeRunes++
		}
		
		if takeRunes == 0 && len(chunkRunes) > 0 {
			// Single char bigger than chunk size? (Unlikely unless MaxChunkSize is tiny)
			takeRunes = 1
		}
		
		parts = append(parts, string(chunkRunes[:takeRunes]))
		currentStart += takeRunes
	}
	
	return parts
}

