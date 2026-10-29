CREATE TABLE IF NOT EXISTS event_members (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'listener' CHECK (role IN ('host', 'co_host', 'speaker', 'listener')),
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'left')),
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, user_id) 
);

CREATE INDEX IF NOT EXISTS idx_event_members_event_id ON event_members(event_id);