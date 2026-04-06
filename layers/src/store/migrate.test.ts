import Database from "better-sqlite3";
import { afterEach, describe, expect, it } from "vitest";
import { migrate } from "./migrate.js";

describe("migrate", () => {
  let db: Database.Database;

  afterEach(() => {
    db?.close();
  });

  it("is idempotent for embedding column when user_version is 1 but column already exists", () => {
    db = new Database(":memory:");
    db.exec(`
      CREATE TABLE memories (
        id TEXT PRIMARY KEY NOT NULL,
        type TEXT NOT NULL,
        content TEXT NOT NULL,
        category TEXT,
        feature_id INTEGER,
        source TEXT NOT NULL,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        legacy_id TEXT UNIQUE,
        embedding BLOB
      );
    `);
    db.pragma("user_version = 1");

    expect(() => migrate(db)).not.toThrow();
    expect(db.pragma("user_version", { simple: true })).toBe(2);
  });
});
