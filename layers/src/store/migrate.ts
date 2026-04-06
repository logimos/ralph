import type Database from "better-sqlite3";

const CURRENT_VERSION = 2;

function tableHasColumn(db: Database.Database, table: string, column: string): boolean {
  const rows = db.prepare(`PRAGMA table_info(${table})`).all() as Array<{ name: string }>;
  return rows.some((r) => r.name === column);
}

export function migrate(db: Database.Database): void {
  let v = db.pragma("user_version", { simple: true }) as number;
  if (v >= CURRENT_VERSION) {
    return;
  }
  if (v === 0) {
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
        legacy_id TEXT UNIQUE
      );

      CREATE INDEX idx_memories_feature_id ON memories(feature_id);
      CREATE INDEX idx_memories_type ON memories(type);

      CREATE VIRTUAL TABLE memory_fts USING fts5(
        memory_id UNINDEXED,
        body,
        tokenize = 'porter unicode61'
      );

      CREATE TRIGGER memories_ai AFTER INSERT ON memories BEGIN
        INSERT INTO memory_fts(memory_id, body) VALUES (new.id, new.content);
      END;
      CREATE TRIGGER memories_au AFTER UPDATE ON memories BEGIN
        DELETE FROM memory_fts WHERE memory_id = old.id;
        INSERT INTO memory_fts(memory_id, body) VALUES (new.id, new.content);
      END;
      CREATE TRIGGER memories_ad AFTER DELETE ON memories BEGIN
        DELETE FROM memory_fts WHERE memory_id = old.id;
      END;
    `);
    v = 1;
  }
  if (v === 1) {
    if (!tableHasColumn(db, "memories", "embedding")) {
      db.exec(`ALTER TABLE memories ADD COLUMN embedding BLOB;`);
    }
    v = 2;
  }
  db.pragma(`user_version = ${CURRENT_VERSION}`);
}
