// @ts-check

/**
 * Helpers for "templatable" safe-output config fields.
 *
 * A templatable field is one that:
 *  - Does NOT affect the generated .lock.yml (no compile-time structural
 *    impact).
 *  - Can be supplied as a literal boolean/string value OR as a GitHub
 *    Actions expression ("${{ inputs.foo }}") that is resolved at runtime
 *    when the env-var containing the handler-config JSON is expanded.
 *
 * The Go counterpart lives in pkg/workflow/templatables.go.
 */

/**
 * Parses a templatable boolean config value.
 *
 * Handles all representations that can arrive in a handler config:
 *  - `undefined` / `null`  → `defaultValue`
 *  - boolean `true`        → `true`
 *  - boolean `false`       → `false`
 *  - string `"true"`       → `true`
 *  - string `"false"`      → `false`
 *  - any other string (e.g. a resolved GitHub Actions expression value
 *    that was not "false") → `true`
 *
 * @param {any} value - The config field value to parse.
 * @param {boolean} [defaultValue=true] - Value returned when `value` is
 *   `undefined` or `null`.
 * @returns {boolean}
 */
function parseBoolTemplatable(value, defaultValue = true) {
  if (value === undefined || value === null) return defaultValue;
  return String(value) !== "false";
}

module.exports = { parseBoolTemplatable };
