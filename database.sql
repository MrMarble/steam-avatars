CREATE TABLE IF NOT EXISTS "users" (
    "id" INTEGER PRIMARY KEY,
    "display_name" TEXT NOT NULL,
    "vanity_url" TEXT,
    "avatar" TEXT,
    "frame" TEXT,
    "created_at" TEXT NOT NULL,
    "updated_at" TEXT
);
