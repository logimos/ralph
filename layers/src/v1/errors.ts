export function errorCodeFrom(e: unknown, fallback: string): string {
  if (e && typeof e === "object" && "code" in e) {
    const c = (e as { code: unknown }).code;
    if (typeof c === "string" && c.length > 0) {
      return c;
    }
  }
  return fallback;
}
