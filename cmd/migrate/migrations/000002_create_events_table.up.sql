CREATE TABLE IF NOT EXISTS categories (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    category_id  INT REFERENCES categories(id) ON DELETE SET NULL,
    visibility   TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    status       TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'live', 'ended', 'cancelled')),
    scheduled_at TIMESTAMPTZ,
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ,
    room_name    TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_status_scheduled ON events (status, scheduled_at);
CREATE INDEX idx_events_visibility ON events (visibility);

-- Seed some default categories
INSERT INTO categories (name, slug) VALUES
('Technology', 'technology'),
('Business', 'business'),
('Education', 'education'),
('Health', 'health')
ON CONFLICT (slug) DO NOTHING;