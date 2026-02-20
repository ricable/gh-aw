package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/github/gh-aw/pkg/logger"
)

var serverLog = logger.New("lsp:server")

// Server is the LSP server that handles JSON-RPC messages over stdio.
type Server struct {
	reader *bufio.Reader
	writer io.Writer
	stderr io.Writer

	docs   *DocumentStore
	schema *SchemaProvider

	shutdown bool
}

// NewServer creates a new LSP server.
func NewServer(stdin io.Reader, stdout io.Writer, stderr io.Writer) (*Server, error) {
	sp, err := NewSchemaProvider()
	if err != nil {
		return nil, fmt.Errorf("initializing schema provider: %w", err)
	}

	return &Server{
		reader: bufio.NewReader(stdin),
		writer: stdout,
		stderr: stderr,
		docs:   NewDocumentStore(),
		schema: sp,
	}, nil
}

// Run starts the LSP server main loop. It reads messages from stdin and writes responses to stdout.
// It returns nil on clean exit (after shutdown+exit) or an error.
func (s *Server) Run() error {
	serverLog.Print("LSP server starting")

	for {
		msg, err := ReadMessage(s.reader)
		if err != nil {
			if s.shutdown {
				return nil
			}
			return fmt.Errorf("reading message: %w", err)
		}

		if err := s.handleMessage(msg); err != nil {
			serverLog.Printf("Error handling %s: %v", msg.Method, err)
		}

		// Exit after the exit notification
		if msg.Method == "exit" {
			return nil
		}
	}
}

func (s *Server) handleMessage(msg *JSONRPCMessage) error {
	serverLog.Printf("Received: method=%s, id=%v", msg.Method, msg.ID)

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return nil // No action needed
	case "shutdown":
		return s.handleShutdown(msg)
	case "exit":
		return nil // Handled by Run loop
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	default:
		// For unknown methods with an ID, send method not found error
		if msg.ID != nil {
			return s.sendError(msg.ID, -32601, "Method not found: "+msg.Method)
		}
		return nil
	}
}

func (s *Server) handleInitialize(msg *JSONRPCMessage) error {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: TextDocumentSyncKindFull,
			HoverProvider:    true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{":", " ", "\n"},
			},
		},
		ServerInfo: &ServerInfo{
			Name:    "gh-aw-lsp",
			Version: "0.1.0",
		},
	}

	return s.sendResponse(msg.ID, result)
}

func (s *Server) handleShutdown(msg *JSONRPCMessage) error {
	s.shutdown = true
	return s.sendResponse(msg.ID, nil)
}

func (s *Server) handleDidOpen(msg *JSONRPCMessage) error {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return fmt.Errorf("unmarshaling didOpen params: %w", err)
	}

	snap := s.docs.Open(
		params.TextDocument.URI,
		params.TextDocument.Version,
		params.TextDocument.Text,
	)

	return s.publishDiagnostics(snap)
}

func (s *Server) handleDidChange(msg *JSONRPCMessage) error {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return fmt.Errorf("unmarshaling didChange params: %w", err)
	}

	// Full sync: use the last content change
	if len(params.ContentChanges) == 0 {
		return nil
	}

	text := params.ContentChanges[len(params.ContentChanges)-1].Text
	snap := s.docs.Update(
		params.TextDocument.URI,
		params.TextDocument.Version,
		text,
	)

	return s.publishDiagnostics(snap)
}

func (s *Server) handleDidClose(msg *JSONRPCMessage) error {
	var params DidCloseTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return fmt.Errorf("unmarshaling didClose params: %w", err)
	}

	s.docs.Close(params.TextDocument.URI)

	// Clear diagnostics for the closed document
	return s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []Diagnostic{},
	})
}

func (s *Server) handleHover(msg *JSONRPCMessage) error {
	var params TextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return fmt.Errorf("unmarshaling hover params: %w", err)
	}

	snap := s.docs.Get(params.TextDocument.URI)
	hover := HandleHover(snap, params.Position, s.schema)

	return s.sendResponse(msg.ID, hover)
}

func (s *Server) handleCompletion(msg *JSONRPCMessage) error {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return fmt.Errorf("unmarshaling completion params: %w", err)
	}

	snap := s.docs.Get(params.TextDocument.URI)
	list := HandleCompletion(snap, params.Position, s.schema)

	return s.sendResponse(msg.ID, list)
}

func (s *Server) publishDiagnostics(snap *DocumentSnapshot) error {
	diags := ComputeDiagnostics(snap)

	return s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         snap.URI,
		Diagnostics: diags,
	})
}

func (s *Server) sendResponse(id any, result any) error {
	msg := &JSONRPCMessage{
		ID:     id,
		Result: result,
	}
	return WriteMessage(s.writer, msg)
}

func (s *Server) sendError(id any, code int, message string) error {
	msg := &JSONRPCMessage{
		ID: id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	return WriteMessage(s.writer, msg)
}

func (s *Server) sendNotification(method string, params any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshaling notification params: %w", err)
	}
	msg := &JSONRPCMessage{
		Method: method,
		Params: raw,
	}
	return WriteMessage(s.writer, msg)
}
