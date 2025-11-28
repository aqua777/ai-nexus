package textsplitter

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParagraphSplitterTestSuite struct {
	suite.Suite
}

func TestParagraphSplitterTestSuite(t *testing.T) {
	suite.Run(t, new(ParagraphSplitterTestSuite))
}

func (s *ParagraphSplitterTestSuite) TestSplitText_Basic() {
	splitter := NewParagraphSplitter(20) // Small chunk size to force splits
	text := "Para1\n\nPara2\n\nPara3"
	
	// Para1 (5) -> chunk: "Para1"
	// Para2 (5) -> chunk + "\n" + "Para2" = 11 <= 20.
	// Para3 (5) -> chunk + "\n" + "Para3" = 17 <= 20.
	// Wait, 20 is limit.
	
	// Let's try specific scenario to test overlap logic.
	// Chunk 1: Para1, Para2
	// Chunk 2: Para2, Para3 (Overlap!)
	
	splitter = NewParagraphSplitter(12)
	// "Para1" (5)
	// + "Para2" (5) + sep (1) = 11. Fits.
	// + "Para3" (5) + sep (1) = 17. No fit.
	// Chunk 1: "Para1\nPara2"
	// Start new Chunk: Overlap "Para2"
	// + "Para3" (5) + sep (1) = 11. Fits.
	
	chunks := splitter.SplitText(text)
	s.Equal(2, len(chunks))
	s.Equal("Para1\nPara2", chunks[0])
	s.Equal("Para2\nPara3", chunks[1])
}

func (s *ParagraphSplitterTestSuite) TestSplitText_Oversized() {
	splitter := NewParagraphSplitter(5)
	text := "1234567890"
	
	chunks := splitter.SplitText(text)
	// Should be split into "12345", "67890"
	s.Equal(2, len(chunks))
	s.Equal("12345", chunks[0])
	s.Equal("67890", chunks[1])
}

