import Database from "better-sqlite3";
import { dbPath } from "./paths.js";
import { migrate } from "./migrate.js";

export function openDatabase(dataDir: string): Database.Database {
  const path = dbPath(dataDir);
  const db = new Database(path);
  db.pragma("journal_mode = WAL");
  db.pragma("foreign_keys = ON");
  migrate(db);
  return db;
}
