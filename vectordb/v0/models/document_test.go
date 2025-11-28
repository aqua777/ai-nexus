package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DocumentTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *DocumentTestSuite) SetupSuite() {
	s.tempDir = s.T().TempDir()
}

func TestDocumentTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentTestSuite))
}

func (s *DocumentTestSuite) TestDocumentFromFile_Success() {
	// Create a temporary file
	filename := "test_document.txt"
	filePath := filepath.Join(s.tempDir, filename)
	content := "This is a test document content."
	err := os.WriteFile(filePath, []byte(content), 0644)
	s.Require().NoError(err)

	// Call the function
	doc, err := DocumentFromFile(filePath)

	// Assertions
	s.Require().NoError(err)
	s.Require().NotNil(doc)
	s.NotEmpty(doc.ID, "ID should not be empty")
	s.Equal(filename, doc.Title, "Title should match filename")
	s.Equal(content, doc.Content, "Content should match file content")
	s.Empty(doc.Metadata, "Metadata should be empty")
}

func (s *DocumentTestSuite) TestDocumentFromFile_FileNotFound() {
	nonExistentPath := filepath.Join(s.tempDir, "non_existent_file.txt")

	doc, err := DocumentFromFile(nonExistentPath)

	s.Error(err)
	s.Nil(doc)
}

