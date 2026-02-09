---
"gh-aw": patch
---

Prevent race conditions in the `add_comment` tool by isolating handler state inside `batchContext`, ensuring each batch maintains its own processed counts and temporary ID mappings.
