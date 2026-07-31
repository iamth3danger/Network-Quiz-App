-- name: CreateQuestion :one
INSERT INTO questions (objective_id, question, a, b, c, d, answer, term_id)
       VALUES ($1,$2,$3,$4,$5,$6,$7,$8) 
       RETURNING *;

-- name: GetReviewedQuestions :many
SELECT q.*, s.* FROM questions q
                  JOIN scores s ON s.question_id = q.id
                  AND q.objective_id = $1;

-- name: GetNewQuestions :many 
SELECT q.id, q.question, q.a, q.b, q.c, q.d, q.answer, q.objective_id, q.term_id
FROM questions q
LEFT JOIN scores s ON s.question_id = q.id
WHERE s.question_id IS NULL
AND q.objective_id = $1;


-- name: GetQuestions :many
SELECT q.id, q.objective_id, s.question_id, s.right_count, s.wrong_count, s.streak 
        FROM questions q 
        LEFT JOIN scores s ON s.question_id = q.id 
        WHERE q.objective_id = $1;

-- name: GetQuestionSetByExamId :many 
SELECT question FROM questions 
                WHERE objective_id = $1;

-- name: GetQuestionsByIds :many
SELECT question FROM questions 
                WHERE term_id = ANY(@ids::int[]);

-- name: GetQuestionCount :one
SELECT COUNT(q.*) FROM questions q
                  JOIN scores s ON s.question_id = q.id
                  WHERE s.right_count <= 2 * s.wrong_count
                  AND s.streak < 3 
                  AND q.objective_id = $1;

-- name: CreateScore :one
INSERT INTO scores (question_id, right_count, wrong_count, streak) VALUES ($1,$2,$3,$4) RETURNING *;

-- name: GetScoreByQuestionId :one
SELECT * FROM scores WHERE question_id = $1;

-- name: UpdateRightScore :one
UPDATE scores SET right_count = right_count + 1, streak = streak + 1 WHERE question_id = $1 RETURNING *;

-- name: UpdateWrongScore :one
UPDATE scores SET wrong_count = wrong_count + 1, streak = 0 WHERE question_id = $1 RETURNING *;

-- name: CreateTurn :one
INSERT INTO turns (question_id, turn) VALUES ($1,$2) RETURNING *;

-- name: CreateOrUpdateTurn :exec
INSERT INTO turns (question_id, turn) VALUES ($1, $2) ON CONFLICT (question_id) DO UPDATE SET turn = $2;

-- name: GetCurrentTurn :one 
SELECT turn FROM state WHERE id = 1;

-- name: UpdateTurn :one
UPDATE state SET turn = turn + 1 WHERE id = 1 RETURNING turn;

-- name: CreateExamObj :exec
INSERT INTO exam_objectives(code, description) VALUES ($1, $2);

-- name: GetExamIdByCode :one
SELECT id FROM exam_objectives WHERE code = $1;

-- name: GetExamCodes :many 
SELECT code FROM exam_objectives;

-- name: GetExamIdsCodes :many 
SELECT id, code FROM exam_objectives;

-- name: CreateObjectiveScore :one
INSERT INTO objective_scores (objective_id) VALUES ($1) RETURNING *;

-- name: UpdateObjectiveScoreRight :one 
UPDATE objective_scores SET right_count = right_count + 1 WHERE objective_id = $1 RETURNING *;

-- name: UpdateObjectiveScoreWrong :one 
UPDATE objective_scores SET wrong_count = wrong_count + 1 WHERE objective_id = $1 RETURNING *;

-- name: CreateQuizQuestion :one
INSERT INTO quiz (question_id, objective_id, status) VALUES ($1, $2, $3) RETURNING *;

-- name: BatchQuizQuestion :exec 
INSERT INTO quiz (question_id, objective_id, status) VALUES (
    unnest(@question_ids::integer[]), 
    unnest(@objective_ids::integer[]), 
    unnest(@status::text[])::question_status
);

-- name: SetQuizStatusPending :one
UPDATE quiz SET status = 'pending' WHERE question_id = $1 RETURNING *;

-- name: SetQuizStatusReviewing :one
UPDATE quiz SET status = 'reviewing' WHERE question_id = $1 RETURNING *;

-- name: SetQuizStatusAnswered :one
UPDATE quiz SET status = 'answered' WHERE question_id = $1 RETURNING *;

-- name: DeleteQuizWithExamId :exec
DELETE FROM quiz
WHERE objective_id = $1;


-- name: GetRandomQuestion :one
SELECT q.id, q.question, q.a, q.b, q.c, q.d, q.answer, q.objective_id, q.term_id
FROM questions q
JOIN quiz qi ON qi.question_id = q.id
WHERE qi.status = 'pending'
AND q.objective_id = $1
ORDER BY RANDOM()
LIMIT 1;

-- name: FetchReviewQuestion :one
SELECT  q.id, q.question, q.a, q.b, q.c, q.d, q.answer, q.objective_id, q.term_id
FROM questions q
JOIN quiz qi on qi.question_id = q.id
WHERE qi.status = 'reviewing'
AND q.objective_id = $1
LIMIT 1;

-- name: GetNumberPending :one
SELECT COUNT(*) FROM quiz WHERE status = 'pending';

-- name: GetNumberAnswered :one
SELECT COUNT(*) FROM quiz WHERE status = 'answered' AND objective_id = $1;

-- name: CreateTerm :one
INSERT INTO terms (objective_id, term_name)
VALUES (@objective_id, @term_name)
ON CONFLICT (term_name) DO NOTHING
RETURNING *;

-- name: GetIdByTerm :one 
SELECT id FROM terms WHERE term_name = $1;

-- name: GetTermsByObjectiveId :many 
SELECT term_name FROM terms WHERE objective_id = $1;

-- name: GetTermQuestionCount :many 
SELECT t.id, t.term_name, COUNT(q.id) as total FROM terms t
LEFT JOIN questions q ON q.term_id = t.id
WHERE t.objective_id = $1
GROUP BY t.id, t.term_name;

-- name: GetTermIds :many
SELECT term_name, id FROM terms;

-- name: GetTerms :many 
SELECT * FROM terms;

-- name: CreateState :one 
INSERT INTO state (turn, objective_id) VALUES ($1,$2) RETURNING *;

-- name: GetExamIds :many
SELECT id FROM exam_objectives;

-- name: TermStats :many 
SELECT e.code, t.term_name, COALESCE(SUM(s.right_count),0) as right_count, COALESCE(SUM(s.wrong_count),0) AS wrong_count FROM questions q JOIN terms t ON t.id = q.term_id LEFT JOIN scores s ON s.question_id = q.id JOIN exam_objectives e ON e.id = q.objective_id GROUP BY t.term_name, e.code ORDER BY e.code ASC;


-- name: GetAllWrongQuestions :many
SELECT q.* FROM questions q 
JOIN scores s on s.question_id = q.id 
WHERE s.wrong_count > 0
AND s.streak = 0;


-- name: GetWrongQuestionsByExam :many
SELECT q.* FROM questions q 
JOIN scores s on s.question_id = q.id 
WHERE s.wrong_count > 0
AND s.streak = 0
AND q.objective_id = $1;

-- name: GetAllExamObjectives :many 
SELECT * FROM exam_objectives;

-- name: GetAllQuestions :many 
SELECT q.* FROM questions q JOIN terms t ON t.id = q.term_id WHERE t.on_exam = true;

-- name: GetAllScores :many 
SELECT * FROM scores;

-- name: GetAllQuiz :many 
SELECT * FROM quiz; 

-- name: GetAllPorts :many 
SELECT * FROM ports;

-- name: GetAllPortScores :many 
SELECT * FROM port_scores;

-- name: GetAllTerms :many 
SELECT * FROM terms;

-- name: GetAllMatching :many
SELECT * FROM matching_quiz;

-- name: GetAllObjectiveScore :many
SELECT * FROM objective_scores;

-- name: CreatePorts :exec 
INSERT INTO ports (abbreviation, port, protocol) VALUES (
    unnest(@abbreviations::text[]), 
    unnest(@ports::integer[]), 
    unnest(@protocols::text[])
);

-- name: UpdatePortScoreRight :exec
INSERT INTO port_scores (port_id, right_count, streak)
VALUES ($1, 1, 1)
ON CONFLICT (port_id)
DO UPDATE SET right_count = port_scores.right_count + 1,
              streak = port_scores.streak + 1;

-- name: UpdatePortScoreWrong :exec
INSERT INTO port_scores (port_id, wrong_count)
VALUES ($1, 1)
ON CONFLICT (port_id)
DO UPDATE SET wrong_count = port_scores.wrong_count + 1,
    streak = 0;

-- name: CreateMatchingQuiz :exec
INSERT INTO matching_quiz (port_id) VALUES (
    unnest(@port_ids::integer[]) 
);

-- name: DeleteMatchingQuiz :exec 
DELETE FROM matching_quiz;

-- name: GetMatchingQuiz :many
SELECT p.id,  p.abbreviation, p.port FROM ports p 
JOIN matching_quiz m ON m.port_id = p.id;

-- name: GetPorts :many 
SELECT p.*, s.* FROM ports p 
LEFT JOIN port_scores s ON s.port_id = p.id;

-- name: GetAverageQuizStreak :one
SELECT AVG(COALESCE(s.streak, 0)) FROM matching_quiz m LEFT JOIN port_scores s ON s.port_id = m.port_id;

-- name: StatsByObjective :many
SELECT
    e.code,
    SUM(COALESCE(s.right_count, 0))                                    AS num_right,
    SUM(COALESCE(s.wrong_count, 0))                                    AS num_wrong,
    COUNT(DISTINCT q.term_id)                                          AS num_terms,
    COUNT(DISTINCT CASE WHEN s.question_id IS NOT NULL THEN q.term_id END) AS num_terms_attempted
FROM questions q JOIN exam_objectives e ON e.id = q.objective_id
LEFT JOIN scores s ON s.question_id = q.id
GROUP BY e.code;

-- name: GetAutoGenScores :many 
SELECT * FROM autogen_scores;

