-- name: CreateGoal :one
INSERT INTO goals (
    id,
    user_id,
    name,
    currency,
    target_amount,
    target_date,
    status,
    created_at,
    updated_at,
    completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetGoalByID :one
SELECT id, user_id, name, currency, target_amount, target_date, status, created_at, updated_at, completed_at 
FROM goals 
WHERE id = $1 
AND user_id = $2 
AND deleted_at IS NULL LIMIT 1;

-- name: ListGoals :many
SELECT id, user_id, name, currency, target_amount, target_date, status, created_at, updated_at, completed_at 
FROM goals 
WHERE user_id = sqlc.arg(user_id) 
AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)) 
ORDER BY created_at DESC 
LIMIT sqlc.arg(page_limit) 
OFFSET sqlc.arg(page_offset);

-- name: UpdateGoal :one
UPDATE goals 
SET 
    name = $3,
    target_amount = $4,
    target_date = $5,
    status = $6,
    updated_at = $7,
    completed_at = $8 
WHERE id = $1 AND user_id = $2;

-- name: DeleteGoal :exec
DELETE FROM goals WHERE id = $1 AND user_id = $2;

-- name: CreateGoalContribution :one
INSERT INTO goal_contributions (
    id,
    goal_id,
    user_id,
    amount,
    contribution_date,
    note,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: ListGoalContributions :many
SELECT id, goal_id, user_id, amount, contribution_date, note, created_at, updated_at 
FROM goal_contributions 
WHERE user_id = $1 AND goal_id = $2 
ORDER BY contribution_date DESC, created_at DESC 
LIMIT $3 
OFFSET $4;

-- name: GetTotalContributedAmount :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total 
FROM goal_contributions 
WHERE user_id = $1 AND goal_id = $2;
