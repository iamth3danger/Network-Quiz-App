CREATE INDEX idx_scores_question_id ON scores(question_id);
CREATE INDEX idx_questions_objective_id ON questions(objective_id);
CREATE INDEX idx_quiz_objective_id ON quiz(objective_id);

CREATE TABLE IF NOT EXISTS exam_objectives (
  id SERIAL PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS subjects (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS terms (
  id SERIAL PRIMARY KEY,
  objective_id INTEGER REFERENCES exam_objectives(id),
  term_name TEXT NOT NULL,
  on_exam boolean NOT NULL DEFAULT false
);

ALTER TABLE terms ADD CONSTRAINT terms_term_name_key UNIQUE (term_name);

CREATE TABLE IF NOT EXISTS questions (
  id SERIAL PRIMARY KEY,
  subject_id INTEGER REFERENCES subjects(id),
  objective_id INTEGER REFERENCES exam_objectives(id),
  question TEXT NOT NULL,
  a TEXT NOT NULL,
  b TEXT NOT NULL,
  c TEXT NOT NULL,
  d TEXT NOT NULL,
  answer TEXT NOT NULL,
  term_id INTEGER REFERENCES terms(id)
);

CREATE TABLE IF NOT EXISTS scores (
  id SERIAL PRIMARY KEY,
  question_id INTEGER REFERENCES questions(id) UNIQUE,
  right_count INTEGER DEFAULT 0,
  wrong_count INTEGER DEFAULT 0,
  streak INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS autogen_scores(
  id SERIAL PRIMARY KEY,
  subject_id INTEGER REFERENCES questions(id) UNIQUE,
  question TEXT NOT NULL,
  right_count INTEGER DEFAULT 0,
  wrong_count INTEGER DEFAULT 0,
  streak INTEGER DEFAULT 0
);

CREATE TYPE question_status AS ENUM ('pending', 'reviewing', 'answered');

CREATE TABLE IF NOT EXISTS quiz (
  id SERIAL PRIMARY KEY,
  question_id INTEGER REFERENCES questions(id) UNIQUE,
  objective_id INTEGER REFERENCES exam_objectives(id),
  subject_id INTEGER REFERENCES subjects(id),
  status question_status
);

CREATE TABLE IF NOT EXISTS objective_scores (
  id SERIAL PRIMARY KEY,
  objective_id INTEGER REFERENCES exam_objectives(id) UNIQUE,
  right_count INTEGER DEFAULT 0,
  wrong_count INTEGER DEFAULT 0
 );

CREATE TABLE IF NOT EXISTS turns (
  id SERIAL PRIMARY KEY,
  question_id INTEGER REFERENCES questions(id) UNIQUE,
  turn INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS state (
  id SERIAL PRIMARY KEY,
  turn INTEGER NOT NULL,
  objective_id INTEGER REFERENCES exam_objectives(id)
);

CREATE TABLE IF NOT EXISTS ports (
   id SERIAL PRIMARY KEY,
   abbreviation TEXT NOT NULL,
   port INTEGER NOT NULL,
   protocol TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS port_scores (
  id SERIAL PRIMARY KEY,
  port_id INTEGER REFERENCES ports(id) UNIQUE,
  right_count INTEGER DEFAULT 0,
  wrong_count INTEGER DEFAULT 0,
  streak INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS matching_quiz (
  id SERIAL PRIMARY KEY,
  port_id INTEGER REFERENCES ports(id) UNIQUE
);
