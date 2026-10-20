# network+ quiz app

quiz app for studying for the COMPTIA Network+ exam

## features

- ports quiz
- matching
- multiple choice questions
- quiz for each exam objective
- tracks per question performance to repeat back worst questions

## stack

[NextJS, Go, PostgreSQL]

## setup

```bash
git clone
docker buildx build -t "network-front" .
cd ser/
docker buildx build -t "network-back" .
```

.env

```bash
POSTGRES_HOST
POSTGRES_PORT
POSTGRES_USER
POSTGRES_PASSWORD
POSTGRES_DB
API_BASE_URL
ANTHROPIC_API_KEY
ANTHROPIC_AUTH_TOKEN
```

To create the database

```bash
docker pull postgres:latest
docker run postgres
psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE quiz;"
psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d quiz -f schema.sql
```

To generate question set with Anthropic API

```bash
go run ./ser/cmd/quest
```
