-- Eventor Mobile App Database Schema
-- Created: 2026-05-12

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
    CREATE TYPE user_role AS ENUM ('user', 'moderator');
  END IF;
END$$;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Users table
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role user_role NOT NULL DEFAULT 'user',
  about TEXT,
  city VARCHAR(128) NOT NULL,
  faculty VARCHAR(255),
  course VARCHAR(10),
  avatar_image_id VARCHAR(128),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Events table
CREATE TABLE IF NOT EXISTS events (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  short_description TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL,
  event_date DATE NOT NULL,
  location VARCHAR(255) NOT NULL,
  image_id VARCHAR(512),
  finished BOOLEAN NOT NULL DEFAULT FALSE,
  creator_id INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TRIGGER IF EXISTS trg_events_updated_at ON events;
CREATE TRIGGER trg_events_updated_at
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_events_creator ON events(creator_id);
CREATE INDEX IF NOT EXISTS idx_events_date ON events(event_date);
CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at);

-- Event tags table
CREATE TABLE IF NOT EXISTS event_tags (
  id SERIAL PRIMARY KEY,
  event_id INTEGER NOT NULL,
  tag VARCHAR(100) NOT NULL,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_event_tags_event ON event_tags(event_id);
CREATE INDEX IF NOT EXISTS idx_event_tags_tag ON event_tags(tag);

-- Event attendees (registration) table
CREATE TABLE IF NOT EXISTS event_attendees (
  id SERIAL PRIMARY KEY,
  event_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT unique_registration UNIQUE (event_id, user_id),
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_event_attendees_user ON event_attendees(user_id);
CREATE INDEX IF NOT EXISTS idx_event_attendees_event ON event_attendees(event_id);

-- Authentication tokens table (for JWT/sessions)
CREATE TABLE IF NOT EXISTS auth_tokens (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  token_hash VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_user ON auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires ON auth_tokens(expires_at);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_users_updated ON users(updated_at);
CREATE INDEX IF NOT EXISTS idx_events_updated ON events(updated_at);
