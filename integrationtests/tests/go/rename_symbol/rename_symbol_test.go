package rename_symbol_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/go/internal"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestRenameSymbol tests the RenameSymbol functionality with the Go language server.
// Runs in both subprocess and headless (listen-mode) modes.
func TestRenameSymbol(t *testing.T) {
	for _, mode := range []struct {
		name     string
		headless bool
	}{{"Subprocess", false}, {"Headless", true}} {
		mode := mode
		snapshotCategory := "rename_symbol"
		if mode.headless {
			snapshotCategory = "rename_symbol_headless"
		}
		t.Run(mode.name, func(t *testing.T) {
			t.Run("SuccessfulRename", func(t *testing.T) {
				suite := internal.GetTestSuiteForMode(t, mode.headless)

				time.Sleep(2 * time.Second)

				ctx, cancel := context.WithTimeout(suite.Context, 5*time.Second)
				defer cancel()

				filePath := filepath.Join(suite.WorkspaceDir, "types.go")
				err := suite.Client.OpenFile(ctx, filePath)
				if err != nil {
					t.Fatalf("Failed to open types.go: %v", err)
				}

				result, err := tools.RenameSymbol(ctx, suite.Client, filePath, 25, 7, "UpdatedConstant")
				if err != nil {
					t.Fatalf("RenameSymbol failed: %v", err)
				}

				if !strings.Contains(result, "Successfully renamed symbol") {
					t.Errorf("Expected success message but got: %s", result)
				}

				if !strings.Contains(result, "occurrences") {
					t.Errorf("Expected multiple occurrences to be renamed but got: %s", result)
				}

				common.SnapshotTest(t, "go", snapshotCategory, "successful", result)

				fileContent, err := suite.ReadFile("types.go")
				if err != nil {
					t.Fatalf("Failed to read types.go: %v", err)
				}

				if !strings.Contains(fileContent, "UpdatedConstant") {
					t.Errorf("Expected to find renamed constant 'UpdatedConstant' in types.go")
				}

				consumerContent, err := suite.ReadFile("consumer.go")
				if err != nil {
					t.Fatalf("Failed to read consumer.go: %v", err)
				}

				if !strings.Contains(consumerContent, "UpdatedConstant") {
					t.Errorf("Expected to find renamed constant 'UpdatedConstant' in consumer.go")
				}
			})

			t.Run("SymbolNotFound", func(t *testing.T) {
				suite := internal.GetTestSuiteForMode(t, mode.headless)

				time.Sleep(2 * time.Second)

				ctx, cancel := context.WithTimeout(suite.Context, 5*time.Second)
				defer cancel()

				filePath := filepath.Join(suite.WorkspaceDir, "clean.go")
				err := suite.Client.OpenFile(ctx, filePath)
				if err != nil {
					t.Fatalf("Failed to open clean.go: %v", err)
				}

				_, err = tools.RenameSymbol(ctx, suite.Client, filePath, 10, 10, "NewName")

				if err == nil {
					t.Errorf("Expected an error when renaming non-existent symbol, but got success")
				}

				errorMessage := err.Error()

				if !strings.Contains(errorMessage, "failed to rename") && !strings.Contains(errorMessage, "column is beyond") {
					t.Errorf("Expected error message about failed rename but got: %s", errorMessage)
				}

				common.SnapshotTest(t, "go", snapshotCategory, "not_found", errorMessage)
			})
		})
	}
}
