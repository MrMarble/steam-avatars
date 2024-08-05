PRAGMA defer_foreign_keys=TRUE;

CREATE TABLE IF NOT EXISTS "users" (
    "id" INTEGER PRIMARY KEY,
    "display_name" TEXT NOT NULL,
    "vanity_url" TEXT,
    "avatar" TEXT,
    "frame" TEXT,
    "created_at" TEXT NOT NULL,
    "updated_at" TEXT
);

CREATE TABLE IF NOT EXISTS "queries" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "query" TEXT NOT NULL,
    "ip" TEXT NOT NULL,
    "country" TEXT NOT NULL,
    "created_at" TEXT NOT NULL
);
