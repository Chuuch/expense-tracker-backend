CREATE TABLE goals (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    currency TEXT NOT NULL,
    target_amount BIGINT NOT NULL,
    target_date DATE,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    CONSTRAINT chk_goals_target_amount_positive CHECK (target_amount > 0)
);

CREATE INDEX idx_goals_user_id ON goals (user_id);
CREATE INDEX idx_goals_user_id_status ON goals (user_id, status);

CREATE TABLE goal_contributions (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    amount BIGINT NOT NULL,
    contribution_date DATE NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_goal_contributions_amount_positive CHECK (amount > 0)
);

CREATE INDEX idx_goal_contributions_goal_id ON goal_contributions (goal_id);
CREATE INDEX idx_goal_contributions_user_goal ON goal_contributions (user_id, goal_id);