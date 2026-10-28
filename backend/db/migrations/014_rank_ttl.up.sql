ALTER TABLE user_rank_discipline
    ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
