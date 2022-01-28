#!/bin/bash

set -e

echo "Wait for database"
/app/wait-for-it.sh db:5432

echo "Migrate database"
/app/migrate -source file://migrations -database postgres://${TX_DB_USER}:${TX_DB_PASSWORD}@${TX_DB_HOST}:${TX_DB_PORT}/${TX_DB_NAME}?sslmode=disable up

echo "Run app"
/app/billionaire