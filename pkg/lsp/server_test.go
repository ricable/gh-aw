//go:build !integration

package lsp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_InitializeShutdownExit(t *testing.T) {
	input := ""
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"processId": nil, "rootUri": "file:///test"},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "shutdown",
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "exit",
	})

	var stdout bytes.Buffer
	server, err := NewServer(strings.NewReader(input), &stdout, &bytes.Buffer{})
	require.NoError(t, err, "NewServer should succeed")

	err = server.Run()
	require.NoError(t, err, "server should exit cleanly")

	// Parse responses
	responses := parseLSPMessages(t, stdout.Bytes())
	require.GreaterOrEqual(t, len(responses), 2, "should have at least 2 responses (initialize + shutdown)")

	// Check initialize response
	initResp := responses[0]
	result := initResp["result"].(map[string]any)
	caps := result["capabilities"].(map[string]any)
	assert.NotNil(t, caps["textDocumentSync"], "should have text document sync")
	assert.Equal(t, true, caps["hoverProvider"], "should have hover provider")
	assert.NotNil(t, caps["completionProvider"], "should have completion provider")
}

func TestServer_DidOpenPublishesDiagnostics(t *testing.T) {
	input := ""
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"processId": nil},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "textDocument/didOpen",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": "file:///test.md", "languageId": "markdown",
				"version": 1, "text": "---\nengine: copilot\n---\n# Title",
			},
		},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 99, "method": "shutdown",
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "exit",
	})

	var stdout bytes.Buffer
	server, err := NewServer(strings.NewReader(input), &stdout, &bytes.Buffer{})
	require.NoError(t, err, "NewServer should succeed")

	err = server.Run()
	require.NoError(t, err, "server should exit cleanly")

	responses := parseLSPMessages(t, stdout.Bytes())

	// Find the diagnostics notification
	var diagnosticsFound bool
	for _, resp := range responses {
		if resp["method"] == "textDocument/publishDiagnostics" {
			diagnosticsFound = true
			params := resp["params"].(map[string]any)
			assert.Equal(t, "file:///test.md", params["uri"], "diagnostics URI should match")
			diags, ok := params["diagnostics"].([]any)
			if ok && len(diags) > 0 {
				// Should have error about missing 'on'
				firstDiag := diags[0].(map[string]any)
				assert.Contains(t, firstDiag["message"].(string), "on", "diagnostic should mention 'on'")
			}
			break
		}
	}
	assert.True(t, diagnosticsFound, "should publish diagnostics on didOpen")
}

func TestServer_HoverReturnsResult(t *testing.T) {
	docText := "---\non:\n  issues:\n    types: [opened]\nengine: copilot\n---\n# Title"

	input := ""
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"processId": nil},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "textDocument/didOpen",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": "file:///test.md", "languageId": "markdown",
				"version": 1, "text": docText,
			},
		},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "textDocument/hover",
		"params": map[string]any{
			"textDocument": map[string]any{"uri": "file:///test.md"},
			"position":     map[string]any{"line": 4, "character": 2},
		},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 99, "method": "shutdown",
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "exit",
	})

	var stdout bytes.Buffer
	server, err := NewServer(strings.NewReader(input), &stdout, &bytes.Buffer{})
	require.NoError(t, err, "NewServer should succeed")

	err = server.Run()
	require.NoError(t, err, "server should exit cleanly")

	responses := parseLSPMessages(t, stdout.Bytes())

	// Find hover response (id=2)
	var hoverResp map[string]any
	for _, resp := range responses {
		id, ok := resp["id"]
		if ok && id != nil {
			idNum, isNum := id.(float64)
			if isNum && int(idNum) == 2 {
				hoverResp = resp
				break
			}
		}
	}
	require.NotNil(t, hoverResp, "should have hover response")
	result := hoverResp["result"].(map[string]any)
	contents := result["contents"].(map[string]any)
	assert.Equal(t, "markdown", contents["kind"], "hover should be markdown")
	assert.Contains(t, contents["value"].(string), "engine", "hover should mention engine")
}

func TestServer_UnknownMethodReturnsError(t *testing.T) {
	input := ""
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"processId": nil},
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "textDocument/unknownMethod",
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "id": 3, "method": "shutdown",
	})
	input += buildLSPMsg(t, map[string]any{
		"jsonrpc": "2.0", "method": "exit",
	})

	var stdout bytes.Buffer
	server, err := NewServer(strings.NewReader(input), &stdout, &bytes.Buffer{})
	require.NoError(t, err, "NewServer should succeed")

	err = server.Run()
	require.NoError(t, err, "server should exit cleanly")

	responses := parseLSPMessages(t, stdout.Bytes())

	// Find error response (id=2)
	var errorResp map[string]any
	for _, resp := range responses {
		id, ok := resp["id"]
		if ok && id != nil {
			idNum, isNum := id.(float64)
			if isNum && int(idNum) == 2 {
				errorResp = resp
				break
			}
		}
	}
	require.NotNil(t, errorResp, "should have error response for unknown method")
	errObj := errorResp["error"].(map[string]any)
	code, ok := errObj["code"].(float64)
	require.True(t, ok, "error code should be a number")
	assert.Equal(t, -32601, int(code), "should be method not found error")
}

// Helper functions

func buildLSPMsg(t *testing.T, msg map[string]any) string {
	t.Helper()
	body, err := json.Marshal(msg)
	require.NoError(t, err, "marshaling test message")
	return "Content-Length: " + json.Number(func() string {
		b, _ := json.Marshal(len(body))
		return string(b)
	}()).String() + "\r\n\r\n" + string(body)
}

func parseLSPMessages(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var messages []map[string]any
	pos := 0
	prefix := []byte("Content-Length: ")

	for pos < len(data) {
		cl := bytes.Index(data[pos:], prefix)
		if cl == -1 {
			break
		}
		cl += pos

		hdrEnd := bytes.Index(data[cl:], []byte("\r\n\r\n"))
		if hdrEnd == -1 {
			break
		}
		hdrEnd += cl

		lengthStr := string(data[cl+len(prefix) : hdrEnd])
		var length int
		err := json.Unmarshal([]byte(lengthStr), &length)
		require.NoError(t, err, "parsing content length")

		bodyStart := hdrEnd + 4
		body := data[bodyStart : bodyStart+length]

		var msg map[string]any
		err = json.Unmarshal(body, &msg)
		require.NoError(t, err, "parsing response body")

		messages = append(messages, msg)
		pos = bodyStart + length
	}

	return messages
}
