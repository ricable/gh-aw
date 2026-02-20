package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// JSONRPCMessage is a JSON-RPC 2.0 message.
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ReadMessage reads a single LSP message from the reader using Content-Length framing.
func ReadMessage(reader *bufio.Reader) (*JSONRPCMessage, error) {
	var contentLength int

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("reading header: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		if strings.HasPrefix(line, "Content-Length: ") {
			val := strings.TrimPrefix(line, "Content-Length: ")
			contentLength, err = strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// Read body
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var msg JSONRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("unmarshaling message: %w", err)
	}

	return &msg, nil
}

// WriteMessage writes an LSP message to the writer with Content-Length framing.
func WriteMessage(writer io.Writer, msg *JSONRPCMessage) error {
	msg.JSONRPC = "2.0"
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(writer, header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	if _, err := writer.Write(body); err != nil {
		return fmt.Errorf("writing body: %w", err)
	}

	return nil
}
