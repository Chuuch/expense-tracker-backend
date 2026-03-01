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
WHERE user_id = $1 
AND ($2::text IS NULL OR category = $2) 
AND ($3::timestamptz IS NULL OR date >= $3) 
AND ($4::timestamptz IS NULL OR date <= $4) 
ORDER BY date DESC, created_at DESC 
LIMIT $5 OFFSET $6;

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