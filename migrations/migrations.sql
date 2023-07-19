CREATE DATABASE brb;

CREATE TABLE IF NOT EXISTS users (
  id TEXT NOT NULL PRIMARY KEY,
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

CREATE TABLE IF NOT EXISTS messages (
  id           TEXT NOT NULL PRIMARY KEY,
  content      TEXT NOT NULL,
  receiver_id  TEXT NOT NULL REFERENCES users(id),
  sender_id    TEXT NOT NULL REFERENCES users(id),
  created_at   TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  is_deleted   BOOLEAN DEFAULT FALSE,
  is_edited    BOOLEAN DEFAULT FALSE,
  reply_to_id  TEXT REFERENCES messages(id)
);

-- seed data
INSERT INTO users VALUES
('1', 'Zoe', 'zoe@gmail.com', true, 'https://i.pinimg.com/564x/d1/ad/78/d1ad7851db8d995a0b6cc13ce468c4e0.jpg', ''),
('2', 'Zojo', 'zojo@gmail.com', true, '', 'https://i.pinimg.com/564x/fb/e4/f5/fbe4f52b442d7498dd417442c05cc8c3.jpg'),
('3', 'Bishh', 'bishh@gmail.com', true, 'https://i.pinimg.com/564x/8f/ba/b5/8fbab5ee0fe8cfef7d9f25fb76bd2ed0.jpg', ''),
('4', 'Zanzy', 'zanzy@gmail.com', true, 'https://i.pinimg.com/564x/f6/77/76/f677762bd3c2a5015c0d5e3da92c61c5.jpg', ''),
('5', 'Sam', 'sam@gmail.com', true, 'https://i.pinimg.com/564x/e1/4a/0f/e14a0f349b24df4e3040ccb87b9238c0.jpg', ''),
('6', 'Apple', 'apple@gmail.com', true, 'https://i.pinimg.com/564x/cd/60/44/cd60449398a001dfcd5e578da5ebd699.jpg', ''),
('7', 'LiSa', 'lisa@gmail.com', true, '', 'https://i.pinimg.com/564x/0c/b4/51/0cb4510a74d1f1a0d3878cea8c2a8016.jpg'),
('8', 'Phoung', 'phoung@gmail.com', true, 'https://i.pinimg.com/564x/5b/fa/d8/5bfad892faf0b002a9cc9cf8ab9e98a3.jpg', ''),
('9', 'Black', 'black@gmail.com', true, 'https://i.pinimg.com/564x/c9/c5/19/c9c5199373215f3ef4ba4e9c9b20ef63.jpg', '');
