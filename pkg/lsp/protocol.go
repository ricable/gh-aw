package lsp

// LSP protocol types for the agentic workflow language server.
// Minimal hand-rolled types to reduce dependency surface.

// DocumentURI is a URI for a document.
type DocumentURI string

// Position in a text document expressed as zero-based line and character offset.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range in a text document expressed as start and end positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource.
type Location struct {
	URI   DocumentURI `json:"uri"`
	Range Range       `json:"range"`
}

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
}

// TextDocumentItem represents an item to transfer a text document from the client to the server.
type TextDocumentItem struct {
	URI        DocumentURI `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

// VersionedTextDocumentIdentifier is a text document identifier with version.
type VersionedTextDocumentIdentifier struct {
	URI     DocumentURI `json:"uri"`
	Version int         `json:"version"`
}

// TextDocumentContentChangeEvent describes textual changes to a text document.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// TextDocumentPositionParams is a parameter literal used in requests to pass a text document and a position inside that document.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DidOpenTextDocumentParams is the params for textDocument/didOpen notification.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams is the params for textDocument/didChange notification.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseTextDocumentParams is the params for textDocument/didClose notification.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// PublishDiagnosticsParams is the params for textDocument/publishDiagnostics notification.
type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Hover is the result of a hover request.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents a string value which content can be represented in different formats.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// CompletionItemKind describes the kind of a completion entry.
type CompletionItemKind int

const (
	CompletionItemKindText     CompletionItemKind = 1
	CompletionItemKindKeyword  CompletionItemKind = 14
	CompletionItemKindSnippet  CompletionItemKind = 15
	CompletionItemKindValue    CompletionItemKind = 12
	CompletionItemKindProperty CompletionItemKind = 10
	CompletionItemKindEnum     CompletionItemKind = 13
)

// InsertTextFormat defines whether the insert text in a completion item is to be interpreted as plain text or a snippet.
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// CompletionItem represents a completion item.
type CompletionItem struct {
	Label            string             `json:"label"`
	Kind             CompletionItemKind `json:"kind,omitempty"`
	Detail           string             `json:"detail,omitempty"`
	Documentation    *MarkupContent     `json:"documentation,omitempty"`
	InsertText       string             `json:"insertText,omitempty"`
	InsertTextFormat InsertTextFormat    `json:"insertTextFormat,omitempty"`
	SortText         string             `json:"sortText,omitempty"`
	Deprecated       bool               `json:"deprecated,omitempty"`
}

// CompletionList represents a collection of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionParams is the params for textDocument/completion request.
type CompletionParams struct {
	TextDocumentPositionParams
}

// TextDocumentSyncKind defines how the host (editor) should sync document changes.
type TextDocumentSyncKind int

const (
	TextDocumentSyncKindNone        TextDocumentSyncKind = 0
	TextDocumentSyncKindFull        TextDocumentSyncKind = 1
	TextDocumentSyncKindIncremental TextDocumentSyncKind = 2
)

// InitializeParams is the params for the initialize request.
type InitializeParams struct {
	ProcessID    *int   `json:"processId"`
	RootURI      string `json:"rootUri,omitempty"`
	Capabilities any    `json:"capabilities,omitempty"`
}

// ServerCapabilities defines the capabilities of the language server.
type ServerCapabilities struct {
	TextDocumentSync   TextDocumentSyncKind `json:"textDocumentSync"`
	HoverProvider      bool                 `json:"hoverProvider"`
	CompletionProvider *CompletionOptions   `json:"completionProvider,omitempty"`
}

// CompletionOptions describes options for completion support.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// InitializeResult is the result of the initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerInfo contains information about the server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
