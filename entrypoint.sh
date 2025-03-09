#!/bin/sh

set -e

DB_USER=$(cat /run/secrets/db_user)
DB_PASSWORD=$(cat /run/secrets/db_password)
DB_NAME=$(cat /run/secrets/db_name)

CONNECTION_STRING="user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=db port=5435 sslmode=disable"

exec goose -dir=/db/migrations postgres "$CONNECTION_STRING" up