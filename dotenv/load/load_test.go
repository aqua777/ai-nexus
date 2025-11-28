package load

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DotenvTestSuite struct {
	suite.Suite
	tempDir     string
	projectRoot string
	subDir      string
}

func TestDotenvTestSuite(t *testing.T) {
	suite.Run(t, new(DotenvTestSuite))
}

func (s *DotenvTestSuite) SetupSuite() {
	s.tempDir = s.T().TempDir()

	// Create project root with go.mod
	s.projectRoot = filepath.Join(s.tempDir, "project")
	err := os.MkdirAll(s.projectRoot, 0755)
	s.Require().NoError(err)

	// Create go.mod file to mark project root
	goModPath := filepath.Join(s.projectRoot, "go.mod")
	err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
	s.Require().NoError(err)

	// Create .env file in project root
	envPath := filepath.Join(s.projectRoot, ".env")
	envContent := "TEST_KEY=test_value\nANOTHER_KEY=another_value\n"
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	s.Require().NoError(err)

	// Create subdirectory for testing
	s.subDir = filepath.Join(s.projectRoot, "subdir", "nested")
	err = os.MkdirAll(s.subDir, 0755)
	s.Require().NoError(err)
}

func (s *DotenvTestSuite) TestLoadEnvFile_LocatesEnvFileInProjectRoot() {
	// Save original working directory
	originalWd, err := os.Getwd()
	s.Require().NoError(err)
	defer os.Chdir(originalWd)

	// Change to subdirectory
	err = os.Chdir(s.subDir)
	s.Require().NoError(err)

	// Clear any existing env vars that might interfere
	os.Unsetenv("TEST_KEY")
	os.Unsetenv("ANOTHER_KEY")

	// Load env file from subdirectory - should find .env in project root
	err = Env()
	s.NoError(err)

	// Verify environment variables were loaded
	s.Equal("test_value", os.Getenv("TEST_KEY"))
	s.Equal("another_value", os.Getenv("ANOTHER_KEY"))
}

func (s *DotenvTestSuite) TestLoadEnvFile_WithExplicitWorkingDir() {
	// Clear any existing env vars
	os.Unsetenv("TEST_KEY")
	os.Unsetenv("ANOTHER_KEY")

	// Load env file with explicit working directory
	err := Env(s.subDir)
	s.NoError(err)

	// Verify environment variables were loaded
	s.Equal("test_value", os.Getenv("TEST_KEY"))
	s.Equal("another_value", os.Getenv("ANOTHER_KEY"))
}

func (s *DotenvTestSuite) TestLoadEnvFile_StopsAtGoMod() {
	// Create a directory outside the project (no go.mod)
	outsideDir := filepath.Join(s.tempDir, "outside")
	err := os.MkdirAll(outsideDir, 0755)
	s.Require().NoError(err)

	// Try to load from outside directory - should fail
	err = Env(outsideDir)
	s.Error(err)
	s.Contains(err.Error(), "failed to find .env file")
}
