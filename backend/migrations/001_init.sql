CREATE TABLE IF NOT EXISTS polls (
    id          VARCHAR(12) PRIMARY KEY,
    title       VARCHAR(200) NOT NULL,
    creator_id  VARCHAR(100),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE polls ADD COLUMN IF NOT EXISTS creator_id VARCHAR(100);

CREATE TABLE IF NOT EXISTS poll_options (
    id          VARCHAR(36) PRIMARY KEY,
    poll_id     VARCHAR(12) NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    label       VARCHAR(120) NOT NULL,
    position    INTEGER NOT NULL,
    UNIQUE (poll_id, position)
);

CREATE TABLE IF NOT EXISTS votes (
    poll_id     VARCHAR(12) NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    option_id   VARCHAR(36) NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,
    voter_id    VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS votes_option_id_idx ON votes(option_id);
CREATE INDEX IF NOT EXISTS polls_creator_id_idx ON polls(creator_id);
CREATE INDEX IF NOT EXISTS votes_voter_id_idx ON votes(voter_id);

