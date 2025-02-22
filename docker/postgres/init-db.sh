#!/bin/bash
set -e

echo "Waiting for PostgreSQL..."
sleep 5  # Даем базе времени запуститься

until pg_isready -U "$POSTGRES_USER"; do
  sleep 2
done

echo "Creating database $POSTGRES_DB if it does not exist..."
psql -U "$POSTGRES_USER" -tc "SELECT 1 FROM pg_database WHERE datname = '$POSTGRES_DB'" | grep -q 1 || \
psql -U "$POSTGRES_USER" -c "CREATE DATABASE $POSTGRES_DB;"

echo "Restoring database from init.dump..."
pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" -F c /docker-entrypoint-initdb.d/init.dump

echo "Database restoration complete."
