package lsp

import (
	"strings"
	"sync"
)

// DocumentSnapshot holds an in-memory snapshot of a document.
type DocumentSnapshot struct {
	URI     DocumentURI
	Version int
	Text    string
	Lines   []string

	// Frontmatter region (0-based line numbers)
	FrontmatterStartLine int // line of opening ---
	FrontmatterEndLine   int // line of closing ---
	FrontmatterYAML      string
	HasFrontmatter       bool
}

// DocumentStore manages open document snapshots. Session-only, no persistence.
type DocumentStore struct {
	mu   sync.RWMutex
	docs map[DocumentURI]*DocumentSnapshot
}

// NewDocumentStore creates a new document store.
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[DocumentURI]*DocumentSnapshot),
	}
}

// Open adds or replaces a document in the store.
func (s *DocumentStore) Open(uri DocumentURI, version int, text string) *DocumentSnapshot {
	snap := newSnapshot(uri, version, text)
	s.mu.Lock()
	s.docs[uri] = snap
	s.mu.Unlock()
	return snap
}

// Update replaces a document's text in the store (full sync).
func (s *DocumentStore) Update(uri DocumentURI, version int, text string) *DocumentSnapshot {
	return s.Open(uri, version, text)
}

// Close removes a document from the store.
func (s *DocumentStore) Close(uri DocumentURI) {
	s.mu.Lock()
	delete(s.docs, uri)
	s.mu.Unlock()
}

// Get returns a document snapshot, or nil if not found.
func (s *DocumentStore) Get(uri DocumentURI) *DocumentSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docs[uri]
}

func newSnapshot(uri DocumentURI, version int, text string) *DocumentSnapshot {
	lines := strings.Split(text, "\n")
	snap := &DocumentSnapshot{
		URI:     uri,
		Version: version,
		Text:    text,
		Lines:   lines,
	}
	parseFrontmatterRegion(snap)
	return snap
}

// parseFrontmatterRegion detects the --- delimited frontmatter region.
func parseFrontmatterRegion(snap *DocumentSnapshot) {
	snap.HasFrontmatter = false
	if len(snap.Lines) == 0 {
		return
	}

	// First line must be ---
	if strings.TrimSpace(snap.Lines[0]) != "---" {
		return
	}

	// Find closing ---
	for i := 1; i < len(snap.Lines); i++ {
		if strings.TrimSpace(snap.Lines[i]) == "---" {
			snap.HasFrontmatter = true
			snap.FrontmatterStartLine = 0
			snap.FrontmatterEndLine = i
			// Extract the YAML between delimiters
			yamlLines := snap.Lines[1:i]
			snap.FrontmatterYAML = strings.Join(yamlLines, "\n")
			return
		}
	}
}

// PositionInFrontmatter returns true if the position is within the frontmatter region
// (between the --- delimiters, exclusive of the delimiters themselves).
func (snap *DocumentSnapshot) PositionInFrontmatter(pos Position) bool {
	if !snap.HasFrontmatter {
		return false
	}
	return pos.Line > snap.FrontmatterStartLine && pos.Line < snap.FrontmatterEndLine
}
