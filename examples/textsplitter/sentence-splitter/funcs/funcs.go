package funcs

import (
	"fmt"

	splitter "github.com/aqua777/ai-flow/textsplitter"
)

func GetChunkLengths(chunks []string) []int {
	lengths := make([]int, len(chunks))
	for i, chunk := range chunks {
		lengths[i] = len(chunk)
	}
	return lengths
}

func TestSplitters(text string) {
	fmt.Printf("Original Text Length: %d characters\n", len(text))
	fmt.Println("--------------------------------------------------")

	// 1. Default Splitter (SimpleTokenizer, RegexStrategy)
	fmt.Println("Scenario 1: Default Splitter (SimpleTokenizer, RegexStrategy)")
	splitterDefault := splitter.NewSentenceSplitter(0, 0, nil, nil)
	chunksDefault := splitterDefault.SplitText(text)
	fmt.Printf("Generated %d chunks: %v\n", len(chunksDefault), GetChunkLengths(chunksDefault))
	fmt.Println("--------------------------------------------------")

	// 2. TikToken Splitter
	fmt.Println("Scenario 2: TikToken Tokenizer (gpt-3.5-turbo)")
	tikTokenizer, err := splitter.NewTikTokenTokenizer("gpt-3.5-turbo")
	if err != nil {
		fmt.Printf("Failed to init tiktoken: %v\n", err)
	} else {
		splitterTik := splitter.NewSentenceSplitter(200, 20, tikTokenizer, nil)
		chunksTik := splitterTik.SplitText(text)
		fmt.Printf("Generated %d chunks: %v\n", len(chunksTik), GetChunkLengths(chunksTik))
		// Print first chunk sample
		if len(chunksTik) > 0 {
			fmt.Printf("Sample Chunk 1:\n%s\n", chunksTik[0])
		}
	}
	fmt.Println("--------------------------------------------------")

	// 3. Neurosnap Splitter
	fmt.Println("Scenario 3: Neurosnap Sentence Splitter (Embedded English Data)")
	// We use nil to trigger the default embedded english.json data.
	neuroStrategy, err := splitter.NewNeurosnapSplitterStrategy(nil)
	if err != nil {
		fmt.Printf("Failed to init neurosnap strategy: %v\n", err)
	} else {
		splitterNeuro := splitter.NewSentenceSplitter(200, 20, nil, neuroStrategy)
		chunksNeuro := splitterNeuro.SplitText(text)
		fmt.Printf("Generated %d chunks: %v\n", len(chunksNeuro), GetChunkLengths(chunksNeuro))
		if len(chunksNeuro) > 0 {
			// Print first few chunks to see how sentences are respected
			for i := 0; i < 2 && i < len(chunksNeuro); i++ {
				fmt.Printf("Chunk %d:\n%s\n---\n", i+1, chunksNeuro[i])
			}
		}
	}
	fmt.Println("--------------------------------------------------")
}
