/** Aligns with Porter unicode61 tokenizer: letters, marks, numbers (not ASCII-only). */
const TOKEN_SPLIT = /[^\p{L}\p{M}\p{N}]+/u;

/**
 * Build a safe FTS5 MATCH string from free text: tokens AND-joined (quoted for FTS5).
 */
export function buildFtsMatchQuery(normalizedText: string): string | null {
  const words = normalizedText
    .toLowerCase()
    .split(TOKEN_SPLIT)
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
