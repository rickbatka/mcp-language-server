package diagnostics_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/go/internal"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestDiagnosticsHeadless runs the same diagnostics tests with gopls started in listen mode and connected via NewClientHeadless.
func TestDiagnosticsHeadless(t *testing.T) {
	t.Run("CleanFile", func(t *testing.T) {
		suite := internal.GetHeadlessTestSuite(t)

		ctx, cancel := context.WithTimeout(suite.Context, 5*time.Second)
		defer cancel()

		filePath := filepath.Join(suite.WorkspaceDir, "clean.go")
		result, err := tools.GetDiagnosticsForFile(ctx, suite.Client, filePath, 2, true)
		if err != nil {
			t.Fatalf("GetDiagnosticsForFile failed: %v", err)
		}

		if !strings.Contains(result, "No diagnostics found") {
			t.Errorf("Expected no diagnostics but got: %s", result)
		}

		common.SnapshotTest(t, "go", "diagnostics_headless", "clean", result)
	})

	t.Run("FileWithError", func(t *testing.T) {
		suite := internal.GetHeadlessTestSuite(t)

		time.Sleep(2 * time.Second)

		ctx, cancel := context.WithTimeout(suite.Context, 5*time.Second)
		defer cancel()

		filePath := filepath.Join(suite.WorkspaceDir, "main.go")
		result, err := tools.GetDiagnosticsForFile(ctx, suite.Client, filePath, 2, true)
		if err != nil {
			t.Fatalf("GetDiagnosticsForFile failed: %v", err)
		}

		if strings.Contains(result, "No diagnostics found") {
			t.Errorf("Expected diagnostics but got none")
		}

		if !strings.Contains(result, "unreachable") {
			t.Errorf("Expected unreachable code error but got: %s", result)
		}

		common.SnapshotTest(t, "go", "diagnostics_headless", "unreachable", result)
	})

	t.Run("FileDependency", func(t *testing.T) {
		suite := internal.GetHeadlessTestSuite(t)

		time.Sleep(2 * time.Second)

		ctx, cancel := context.WithTimeout(suite.Context, 5*time.Second)
		defer cancel()

		helperPath := filepath.Join(suite.WorkspaceDir, "helper.go")
		consumerPath := filepath.Join(suite.WorkspaceDir, "consumer.go")

		err := suite.Client.OpenFile(ctx, helperPath)
		if err != nil {
			t.Fatalf("Failed to open helper.go: %v", err)
		}

		err = suite.Client.OpenFile(ctx, consumerPath)
		if err != nil {
			t.Fatalf("Failed to open consumer.go: %v", err)
		}

		time.Sleep(2 * time.Second)

		result, err := tools.GetDiagnosticsForFile(ctx, suite.Client, consumerPath, 2, true)
		if err != nil {
			t.Fatalf("GetDiagnosticsForFile failed: %v", err)
		}

		if !strings.Contains(result, "No diagnostics found") {
			t.Errorf("Expected no diagnostics initially but got: %s", result)
		}

		modifiedHelperContent := `package main

// HelperFunction now requires an int parameter
func HelperFunction(value int) string {
	return "hello world"
}
`
		err = suite.WriteFile("helper.go", modifiedHelperContent)
		if err != nil {
			t.Fatalf("Failed to update helper.go: %v", err)
		}

		helperURI := fmt.Sprintf("file://%s", helperPath)

		err = suite.Client.NotifyChange(ctx, helperPath)
		if err != nil {
			t.Fatalf("Failed to notify change to helper.go: %v", err)
		}

		fileChangeParams := protocol.DidChangeWatchedFilesParams{
			Changes: []protocol.FileEvent{
				{
					URI:  protocol.DocumentUri(helperURI),
					Type: protocol.FileChangeType(protocol.Changed),
				},
			},
		}

		err = suite.Client.DidChangeWatchedFiles(ctx, fileChangeParams)
		if err != nil {
			t.Fatalf("Failed to send DidChangeWatchedFiles: %v", err)
		}

		time.Sleep(3 * time.Second)

		err = suite.Client.CloseFile(ctx, consumerPath)
		if err != nil {
			t.Fatalf("Failed to close consumer.go: %v", err)
		}

		err = suite.Client.OpenFile(ctx, consumerPath)
		if err != nil {
			t.Fatalf("Failed to reopen consumer.go: %v", err)
		}

		time.Sleep(3 * time.Second)

		result, err = tools.GetDiagnosticsForFile(ctx, suite.Client, consumerPath, 2, true)
		if err != nil {
			t.Fatalf("GetDiagnosticsForFile failed after dependency change: %v", err)
		}

		if strings.Contains(result, "No diagnostics found") {
			t.Errorf("Expected diagnostics after dependency change but got none")
		}

		if !strings.Contains(result, "argument") && !strings.Contains(result, "parameter") {
			t.Errorf("Expected error about wrong arguments but got: %s", result)
		}

		common.SnapshotTest(t, "go", "diagnostics_headless", "dependency", result)
	})
}
