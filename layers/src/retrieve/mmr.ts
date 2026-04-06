/**
 * Word-level Jaccard similarity in [0, 1].
 */
export function wordJaccard(a: string, b: string): number {
  const ta = tokenize(a);
  const tb = tokenize(b);
  if (ta.size === 0 && tb.size === 0) {
    return 1;
  }
  if (ta.size === 0 || tb.size === 0) {
    return 0;
  }
  let inter = 0;
  for (const x of ta) {
    if (tb.has(x)) {
      inter++;
    }
  }
  const union = ta.size + tb.size - inter;
  return union === 0 ? 0 : inter / union;
}

/** Keep in sync with ftsQuery TOKEN_SPLIT (unicode letters / marks / numbers). */
const TOKEN_SPLIT = /[^\p{L}\p{M}\p{N}]+/u;

function tokenize(s: string): Set<string> {
  const words = s
    .toLowerCase()
    .split(TOKEN_SPLIT)
    .map((w) => w.trim())
    .filter((w) => w.length >= 2);
  return new Set(words);
}

export type MmrItem = { id: string; relevance: number; text: string };

/**
 * MMR selection: diversity vs relevance (lambda higher = more relevance).
 */
export function mmrSelect(items: MmrItem[], k: number, lambda: number): MmrItem[] {
  const kInt = Number.isFinite(k) ? Math.min(items.length, Math.max(0, Math.floor(k))) : 0;
  if (items.length === 0 || kInt === 0) {
    return [];
  }
  const lam = Math.min(1, Math.max(0, lambda));
  const pool = [...items];
  const selected: MmrItem[] = [];

  while (selected.length < kInt && pool.length > 0) {
    let bestIdx = 0;
    let bestScore = -Infinity;
    for (let i = 0; i < pool.length; i++) {
      const cand = pool[i]!;
      let maxSim = 0;
      for (const s of selected) {
        const sim = wordJaccard(cand.text, s.text);
        if (sim > maxSim) {
          maxSim = sim;
        }
      }
      const mmr = lam * cand.relevance - (1 - lam) * maxSim;
      if (mmr > bestScore) {
        bestScore = mmr;
        bestIdx = i;
      }
    }
    const [next] = pool.splice(bestIdx, 1);
    if (next) {
      selected.push(next);
    }
  }
  return selected;
}
