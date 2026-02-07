// Package internal contains shared helpers for Go tests
package internal

import (
	"path/filepath"
	"testing"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
)

// GetTestSuite returns a test suite for Go language server tests (starts gopls as subprocess)
func GetTestSuite(t *testing.T) *common.TestSuite {
	repoRoot, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatalf("Failed to get repo root: %v", err)
	}

	config := common.LSPTestConfig{
		Name:             "go",
		Command:          "gopls",
		Args:             []string{},
		WorkspaceDir:     filepath.Join(repoRoot, "integrationtests/workspaces/go"),
		InitializeTimeMs: 2000,
	}

	suite := common.NewTestSuite(t, config)
	if err := suite.Setup(); err != nil {
		t.Fatalf("Failed to set up test suite: %v", err)
	}
	t.Cleanup(func() { suite.Cleanup() })
	return suite
}

// GetTestSuiteForMode returns a test suite for the given mode. When headless is true, starts gopls in listen
// mode and connects via NewClientHeadless; otherwise starts gopls as a subprocess.
func GetTestSuiteForMode(t *testing.T, headless bool) *common.TestSuite {
	if headless {
		return GetHeadlessTestSuite(t)
	}
	return GetTestSuite(t)
}

// GetHeadlessTestSuite returns a test suite that starts gopls in listen mode (same Command/Args as GetTestSuite)
// and connects via NewClientHeadless. No external server or GOPLS_HEADLESS_ADDR is required.
func GetHeadlessTestSuite(t *testing.T) *common.TestSuite {
	repoRoot, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatalf("Failed to get repo root: %v", err)
	}

	config := common.LSPTestConfig{
		Name:               "go",
		Command:            "gopls",
		Args:               []string{},
		HeadlessListenArg:  "-listen=127.0.0.1:%d",
		WorkspaceDir:       filepath.Join(repoRoot, "integrationtests/workspaces/go"),
		InitializeTimeMs:   2000,
	}

	suite := common.NewTestSuite(t, config)
	if err := suite.Setup(); err != nil {
		t.Fatalf("Failed to set up headless test suite: %v", err)
	}
	t.Cleanup(func() { suite.Cleanup() })
	return suite
}
