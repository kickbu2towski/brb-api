CREATE DATABASE IF NOT EXISTS brb;

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  email_verified BOOL NOT NULL DEFAULT FALSE,
  avatar TEXT NOT NULL,
  bio TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS tokens (
  hash BYTEA PRIMARY KEY,
  user_id TEXT REFERENCES users (id),
  expiry_time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  scope TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS follow_relations (
  follower_id TEXT REFERENCES users (id),
  following_id TEXT REFERENCES users (id),
  CONSTRAINT pk_follows PRIMARY KEY (follower_id, following_id)
);

CREATE TABLE IF NOT EXISTS dms (
  id SERIAL PRIMARY KEY,
  participants text[] NOT NULL
);

CREATE TABLE IF NOT EXISTS reactions (
  id SERIAL PRIMARY KEY,
  reaction TEXT NOT NULL,
  user_id TEXT REFERENCES users (id),
  message_id TEXT REFERENCES messages(id)
);

CREATE TABLE IF NOT EXISTS messages (
  id           TEXT PRIMARY KEY,
  content      TEXT NOT NULL,
  user_id      TEXT NOT NULL REFERENCES users(id),
  dm_id        INTEGER NOT NULL REFERENCES dms(id),
  created_at   TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  is_deleted   BOOLEAN DEFAULT FALSE,
  is_edited    BOOLEAN DEFAULT FALSE,
  reply_to_id  TEXT REFERENCES messages(id)
);
