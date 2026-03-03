-- name: GetMonthlySpending :many
SELECT to_char(date_trunc('month', date), 'YYYY-MM') AS year_month, 
SUM(amount) AS amount, currency 
FROM expenses 
WHERE user_id = sqlc.arg(user_id)
AND date >= sqlc.arg(from_date) 
AND date < (sqlc.arg(to_date) + INTERVAL '1 month') 
GROUP BY year_month, currency 
ORDER BY year_month ASC;

-- name: GetSpendingByCategory :many
SELECT 
category::text AS category,
SUM(amount) AS amount, currency 
FROM expenses 
WHERE user_id = sqlc.arg(user_id)
AND date >= sqlc.arg(from_date) 
AND date <= sqlc.arg(to_date) 
GROUP BY category, currency 
ORDER BY amount DESC;