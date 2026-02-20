//go:build !integration

package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWriteMessage(t *testing.T) {
	msg := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "initialize",
	}

	// Write
	var buf bytes.Buffer
	err := WriteMessage(&buf, msg)
	require.NoError(t, err, "WriteMessage should succeed")

	// Read back
	reader := bufio.NewReader(&buf)
	got, err := ReadMessage(reader)
	require.NoError(t, err, "ReadMessage should succeed")

	assert.Equal(t, "2.0", got.JSONRPC, "JSONRPC version should match")
	assert.NotNil(t, got.ID, "ID should not be nil")
	assert.Equal(t, "initialize", got.Method, "method should match")
}

func TestReadMessage_MissingContentLength(t *testing.T) {
	input := "\r\n\r\n{}"
	reader := bufio.NewReader(bytes.NewBufferString(input))
	_, err := ReadMessage(reader)
	assert.Error(t, err, "should error on missing Content-Length")
}

func TestWriteMessage_ContentLengthFraming(t *testing.T) {
	msg := &JSONRPCMessage{
		ID:     float64(1),
		Method: "test",
	}

	var buf bytes.Buffer
	err := WriteMessage(&buf, msg)
	require.NoError(t, err, "WriteMessage should succeed")

	output := buf.String()
	assert.Contains(t, output, "Content-Length: ", "should contain Content-Length header")
	assert.Contains(t, output, "\r\n\r\n", "should contain header/body separator")
}

func TestReadWriteMessage_WithParams(t *testing.T) {
	params, _ := json.Marshal(map[string]string{"key": "value"})
	msg := &JSONRPCMessage{
		ID:     float64(42),
		Method: "textDocument/hover",
		Params: params,
	}

	var buf bytes.Buffer
	err := WriteMessage(&buf, msg)
	require.NoError(t, err, "WriteMessage should succeed")

	reader := bufio.NewReader(&buf)
	got, err := ReadMessage(reader)
	require.NoError(t, err, "ReadMessage should succeed")

	assert.NotNil(t, got.ID, "ID should not be nil")
	assert.Equal(t, "textDocument/hover", got.Method, "method should match")
	assert.NotNil(t, got.Params, "params should not be nil")
}
