export function cosineSimilarity(a: number[], b: number[]): number {
  if (a.length !== b.length || a.length === 0) {
    return 0;
  }
  let dot = 0;
  let na = 0;
  let nb = 0;
  for (let i = 0; i < a.length; i++) {
    dot += a[i]! * b[i]!;
    na += a[i]! * a[i]!;
    nb += b[i]! * b[i]!;
  }
  const denom = Math.sqrt(na) * Math.sqrt(nb);
  return denom === 0 ? 0 : dot / denom;
}

export function blobToFloat32Array(blob: Buffer | null): Float32Array | null {
  if (!blob || blob.length === 0) {
    return null;
  }
  return new Float32Array(blob.buffer, blob.byteOffset, blob.length / 4);
}

export function float32ArrayToNumbers(f: Float32Array): number[] {
  return Array.from(f);
}
