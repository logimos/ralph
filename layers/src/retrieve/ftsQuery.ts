/**
 * Build a safe FTS5 MATCH string from free text: alphanumeric tokens, AND-joined.
 */
export function buildFtsMatchQuery(normalizedText: string): string | null {
  const words = normalizedText
    .toLowerCase()
    .split(/[^a-z0-9]+/u)
    .map((w) => w.trim())
    .filter((w) => w.length >= 2);

  if (words.length === 0) {
    return null;
  }

  const escaped = words.map((w) => `"${w.replace(/"/g, '""')}"`);
  return escaped.join(" AND ");
}

export function normalizeQueryText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}
