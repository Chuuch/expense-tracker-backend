-- name: CreateExpense :one
INSERT INTO expenses (
    id,
    user_id,
    amount,
    currency,
    category,
    description,
    date,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetExpenseByID :one
SELECT id, user_id, amount, currency, category, description, date, created_at, updated_at 
FROM expenses 
WHERE id = $1 
AND user_id = $2 
LIMIT 1;

-- name: ListExpenses :many
SELECT id, user_id, amount, currency, category, description, date, created_at, updated_at
FROM expenses
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.narg(category)::text IS NULL OR category = sqlc.narg(category))
  AND (sqlc.narg(from_date)::timestamptz IS NULL OR date >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::timestamptz IS NULL OR date <= sqlc.narg(to_date))
ORDER BY date DESC, created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: UpdateExpense :exec
UPDATE expenses 
SET
 amount = $3,
 currency = $4,
 category = $5,
 description = $6,
 date = $7,
 updated_at = $8 
WHERE id = $1 AND user_id = $2;

-- name: DeleteExpense :exec
DELETE FROM expenses WHERE id = $1 AND user_id = $2;