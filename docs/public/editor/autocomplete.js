/**
 * Frontmatter autocomplete for the gh-aw playground editor.
 *
 * Provides schema-driven suggestions for YAML frontmatter keys and values
 * inside a plain <textarea>.
 */

// ---------------------------------------------------------------------------
// (a) YAML Context Parser
// ---------------------------------------------------------------------------

/**
 * Parse the frontmatter context at the cursor position.
 *
 * Returns null if the cursor is outside the `---` frontmatter block, or:
 * {
 *   path: string[]     — nesting path (e.g. ["tools", "github"])
 *   mode: 'key'|'value' — whether we're completing a key or a value
 *   prefix: string      — text typed so far for filtering
 *   indent: number      — current line indentation (spaces)
 *   siblings: string[]  — keys already present at the same level
 * }
 */
export function parseFrontmatterContext(text, cursorPos) {
  const before = text.substring(0, cursorPos);
  const lines = before.split('\n');
  const currentLineIdx = lines.length - 1;
  const currentLine = lines[currentLineIdx];

  // --- Detect frontmatter boundaries ---
  const allLines = text.split('\n');
  let fmStart = -1;
  let fmEnd = -1;
  for (let i = 0; i < allLines.length; i++) {
    if (allLines[i].trim() === '---') {
      if (fmStart === -1) { fmStart = i; }
      else { fmEnd = i; break; }
    }
  }
  // Cursor must be between the two ---
  if (fmStart === -1 || fmEnd === -1) return null;
  if (currentLineIdx <= fmStart || currentLineIdx >= fmEnd) return null;

  // --- Determine indent of current line ---
  const indentMatch = currentLine.match(/^(\s*)/);
  const indent = indentMatch ? indentMatch[1].length : 0;

  // --- Key vs Value mode ---
  const colonIdx = currentLine.indexOf(':');
  const isComment = currentLine.trimStart().startsWith('#');
  if (isComment) return null;

  let mode, prefix, currentKey = null;
  if (colonIdx === -1 || cursorPos - (before.length - currentLine.length) <= colonIdx) {
    // No colon yet, or cursor is before the colon — completing a key
    mode = 'key';
    prefix = currentLine.trimStart();
    // Remove list marker if present
    if (prefix.startsWith('- ')) prefix = prefix.substring(2);
  } else {
    // Cursor is after the colon — completing a value
    mode = 'value';
    const afterColon = currentLine.substring(colonIdx + 1);
    prefix = afterColon.trim();
    // Extract the key name from before the colon
    const beforeColon = currentLine.substring(0, colonIdx).trim();
    const keyMatch = beforeColon.match(/^(?:-\s+)?([a-zA-Z0-9_-]+)$/);
    if (keyMatch) currentKey = keyMatch[1];
  }

  // --- Build nesting path by walking lines above ---
  const path = [];
  let targetIndent = indent;

  // Walk backward from current line to find parent keys
  for (let i = currentLineIdx - 1; i > fmStart; i--) {
    const line = allLines[i];
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;

    const lineIndentMatch = line.match(/^(\s*)/);
    const lineIndent = lineIndentMatch ? lineIndentMatch[1].length : 0;

    if (lineIndent < targetIndent) {
      // This line is a parent — extract its key
      const keyMatch = trimmed.match(/^([a-zA-Z0-9_-]+)\s*:/);
      if (keyMatch) {
        path.unshift(keyMatch[1]);
        targetIndent = lineIndent;
      }
      if (lineIndent === 0) break;
    }
  }

  // --- Collect sibling keys at the same indent level ---
  const siblings = [];
  for (let i = fmStart + 1; i < fmEnd; i++) {
    if (i === currentLineIdx) continue;
    const line = allLines[i];
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;

    const lineIndentMatch = line.match(/^(\s*)/);
    const lineIndent = lineIndentMatch ? lineIndentMatch[1].length : 0;

    if (lineIndent === indent) {
      // Check this is at the same nesting level by verifying same parent path
      // Simple heuristic: same indent = same level (works for well-formatted YAML)
      const keyMatch = trimmed.match(/^([a-zA-Z0-9_-]+)\s*:/);
      if (keyMatch) {
        // Verify same parent: walk backward from line i to build its path
        const linePath = [];
        let checkIndent = lineIndent;
        for (let j = i - 1; j > fmStart; j--) {
          const pLine = allLines[j];
          const pTrimmed = pLine.trim();
          if (!pTrimmed || pTrimmed.startsWith('#')) continue;
          const pIndentMatch = pLine.match(/^(\s*)/);
          const pIndent = pIndentMatch ? pIndentMatch[1].length : 0;
          if (pIndent < checkIndent) {
            const pKeyMatch = pTrimmed.match(/^([a-zA-Z0-9_-]+)\s*:/);
            if (pKeyMatch) linePath.unshift(pKeyMatch[1]);
            checkIndent = pIndent;
            if (pIndent === 0) break;
          }
        }
        if (arraysEqual(linePath, path)) {
          siblings.push(keyMatch[1]);
        }
      }
    }
  }

  return { path, mode, prefix, indent, siblings, currentKey };
}

function arraysEqual(a, b) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

// ---------------------------------------------------------------------------
// (b) Suggestion Generator
// ---------------------------------------------------------------------------

/**
 * Generate autocomplete suggestions from the data + context.
 *
 * Returns: Array<{ label, detail, desc, snippet, kind }>
 *   label   — display text (key name or enum value)
 *   detail  — type tag (e.g. "string", "object")
 *   desc    — short description
 *   snippet — text to insert
 *   kind    — 'key' or 'value'
 */
export function getSuggestions(data, context) {
  if (!data || !context) return [];

  // Navigate to the correct node in the autocomplete data tree
  let node = data.root;
  for (const segment of context.path) {
    if (!node) break;
    // Current node could be the entry itself or we need its children
    const entry = node[segment];
    if (!entry || !entry.children) {
      node = null;
      break;
    }
    node = entry.children;
  }

  if (!node) return [];

  if (context.mode === 'key') {
    return getKeySuggestions(node, context, data.sortOrder);
  } else {
    return getValueSuggestions(node, context, data.root);
  }
}

function getKeySuggestions(node, context, sortOrder) {
  const results = [];
  const prefix = context.prefix.toLowerCase();
  const siblingSet = new Set(context.siblings);

  for (const [key, entry] of Object.entries(node)) {
    // Filter out already-used siblings
    if (siblingSet.has(key)) continue;

    // Filter by prefix
    if (prefix && !key.toLowerCase().startsWith(prefix)) continue;

    let snippet;
    if (entry.children && !entry.leaf) {
      // Object with children — insert key + newline + indent
      snippet = key + ':\n' + ' '.repeat(context.indent + 2);
    } else if (entry.array) {
      snippet = key + ':\n' + ' '.repeat(context.indent + 2) + '- ';
    } else {
      snippet = key + ': ';
    }

    results.push({
      label: key,
      detail: entry.type || '',
      desc: entry.desc || '',
      snippet,
      kind: 'key',
    });
  }

  // Sort by priority order, then alphabetically
  if (sortOrder && context.path.length === 0) {
    const orderMap = {};
    sortOrder.forEach((k, i) => { orderMap[k] = i; });
    results.sort((a, b) => {
      const ia = orderMap[a.label] ?? 999;
      const ib = orderMap[b.label] ?? 999;
      return ia - ib || a.label.localeCompare(b.label);
    });
  } else {
    results.sort((a, b) => a.label.localeCompare(b.label));
  }

  return results;
}

function getValueSuggestions(node, context, rootNode) {
  if (!context.currentKey) return [];

  const entry = node[context.currentKey];
  if (!entry || !entry.enum) return [];

  const lowerPrefix = context.prefix.toLowerCase();
  return entry.enum
    .map(v => String(v))
    .filter(v => !lowerPrefix || v.toLowerCase().startsWith(lowerPrefix))
    .map(v => ({
      label: v,
      detail: '',
      desc: '',
      snippet: v,
      kind: 'value',
    }));
}

// ---------------------------------------------------------------------------
// (c) Autocomplete Dropdown UI
// ---------------------------------------------------------------------------

export class AutocompleteDropdown {
  constructor(textarea) {
    this.textarea = textarea;
    this.items = [];
    this.activeIndex = 0;
    this.visible = false;
    this.onSelect = null;

    // Create dropdown element
    this.el = document.createElement('div');
    this.el.className = 'autocomplete-dropdown';
    this.el.style.display = 'none';
    this.el.setAttribute('role', 'listbox');

    // Create mirror div for cursor position calculation
    this.mirror = document.createElement('div');
    this.mirror.className = 'autocomplete-mirror';
    this.mirror.style.cssText = `
      position: absolute; visibility: hidden; overflow: hidden;
      white-space: pre-wrap; word-wrap: break-word;
      pointer-events: none;
    `;

    // Append to the editor's parent (position: relative container)
    const container = textarea.parentElement;
    container.appendChild(this.el);
    document.body.appendChild(this.mirror);

    // Dismiss on outside click
    this._onDocClick = (e) => {
      if (!this.el.contains(e.target) && e.target !== this.textarea) {
        this.hide();
      }
    };
    document.addEventListener('mousedown', this._onDocClick);
  }

  show(items, onSelect) {
    if (!items || items.length === 0) {
      this.hide();
      return;
    }

    this.items = items;
    this.activeIndex = 0;
    this.onSelect = onSelect;
    this.visible = true;

    this._render();
    this._position();
    this.el.style.display = 'block';
  }

  hide() {
    this.visible = false;
    this.el.style.display = 'none';
    this.items = [];
  }

  moveUp() {
    if (!this.visible) return;
    this.activeIndex = (this.activeIndex - 1 + this.items.length) % this.items.length;
    this._render();
    this._scrollActiveIntoView();
  }

  moveDown() {
    if (!this.visible) return;
    this.activeIndex = (this.activeIndex + 1) % this.items.length;
    this._render();
    this._scrollActiveIntoView();
  }

  accept() {
    if (!this.visible || this.items.length === 0) return false;
    const item = this.items[this.activeIndex];
    if (item && this.onSelect) {
      this.onSelect(item);
    }
    this.hide();
    return true;
  }

  _render() {
    let html = '';
    for (let i = 0; i < this.items.length; i++) {
      const item = this.items[i];
      const active = i === this.activeIndex ? ' active' : '';
      const detail = item.detail ? `<span class="autocomplete-item-type">${esc(item.detail)}</span>` : '';
      const desc = item.desc ? `<span class="autocomplete-item-desc">${esc(item.desc)}</span>` : '';
      html += `<div class="autocomplete-item${active}" data-index="${i}" role="option">
        <span class="autocomplete-item-key">${esc(item.label)}</span>${detail}${desc}
      </div>`;
    }
    this.el.innerHTML = html;

    // Click handlers
    this.el.querySelectorAll('.autocomplete-item').forEach(el => {
      el.addEventListener('mousedown', (e) => {
        e.preventDefault();
        this.activeIndex = parseInt(el.dataset.index);
        this.accept();
      });
    });
  }

  _scrollActiveIntoView() {
    const active = this.el.querySelector('.autocomplete-item.active');
    if (active) {
      active.scrollIntoView({ block: 'nearest' });
    }
  }

  _position() {
    const ta = this.textarea;
    const style = getComputedStyle(ta);

    // Copy textarea styles to mirror
    const props = [
      'fontFamily', 'fontSize', 'fontWeight', 'lineHeight', 'letterSpacing',
      'wordSpacing', 'textIndent', 'paddingTop', 'paddingRight', 'paddingBottom',
      'paddingLeft', 'borderTopWidth', 'borderRightWidth', 'borderBottomWidth',
      'borderLeftWidth', 'boxSizing', 'tabSize',
    ];
    for (const p of props) {
      this.mirror.style[p] = style[p];
    }
    this.mirror.style.width = ta.clientWidth + 'px';

    // Fill mirror with text up to cursor, measure position
    const text = ta.value.substring(0, ta.selectionStart);
    this.mirror.textContent = text;

    // Add a span to measure cursor position
    const span = document.createElement('span');
    span.textContent = '|';
    this.mirror.appendChild(span);

    const mirrorRect = this.mirror.getBoundingClientRect();
    const spanRect = span.getBoundingClientRect();
    const taRect = ta.getBoundingClientRect();

    // Position relative to textarea's parent container
    const container = ta.parentElement;
    const containerRect = container.getBoundingClientRect();

    const cursorX = spanRect.left - mirrorRect.left;
    const cursorY = spanRect.top - mirrorRect.top;

    const left = cursorX + (taRect.left - containerRect.left) - ta.scrollLeft;
    const top = cursorY + (taRect.top - containerRect.top) - ta.scrollTop + parseInt(style.lineHeight);

    this.el.style.left = Math.max(0, left) + 'px';
    this.el.style.top = Math.min(top, ta.clientHeight - 10) + 'px';
  }

  destroy() {
    document.removeEventListener('mousedown', this._onDocClick);
    this.el.remove();
    this.mirror.remove();
  }
}

function esc(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
