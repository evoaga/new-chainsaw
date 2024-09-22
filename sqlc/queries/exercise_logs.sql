-- Exercise log queries

-- name: LogExercise :exec
INSERT INTO exercise_logs (user_id, exercise_id, reps, weight, additional_weight, exercise_type, bodyweight_id, log_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateExerciseLog :exec
UPDATE exercise_logs
SET
    reps = $3,
    weight = $4,
    additional_weight = $5,
    exercise_type = $6,
    bodyweight_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE
    user_id = $1 AND exercise_id = $2 AND log_date = $8;

-- name: GetExerciseLogs :many
SELECT el.id, el.exercise_id, e.name AS exercise_name, el.reps, el.weight, el.log_date
FROM exercise_logs el
JOIN exercises e ON el.exercise_id = e.id
WHERE el.user_id = $1
ORDER BY el.log_date DESC;

-- name: GetExerciseLogByDetails :one
SELECT id
FROM exercise_logs
WHERE user_id = $1
  AND exercise_id = $2
  AND reps = $3
  AND weight = $4
  AND log_date = $5;

-- name: GetExercisesWithLatestLogDate :many
WITH latest_logs AS (
    SELECT
        exercise_id,
        MAX(log_date) AS latest_log_date
    FROM
        exercise_logs
    WHERE
        user_id = $1
    GROUP BY
        exercise_id
)
SELECT
    e.id,
    e.name,
    el.log_date,
    el.reps,
    el.weight,
    el.bodyweight_id
FROM
    exercises e
JOIN
    exercise_logs el
ON
    e.id = el.exercise_id
JOIN
    latest_logs ll
ON
    el.exercise_id = ll.exercise_id AND el.log_date = ll.latest_log_date
WHERE
    el.user_id = $1;


