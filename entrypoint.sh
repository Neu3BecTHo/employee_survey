#!/bin/sh

# Start PostgreSQL in background
su - postgres -c "pg_ctl start -D /var/lib/postgresql/data -l /var/lib/postgresql/logfile"

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL..."
until pg_isready -U survey_user -d survey_app; do
  sleep 1
done
echo "PostgreSQL is ready"

# Run migrations if they exist
if [ -d /app/migrations ]; then
  echo "Running migrations..."
  for f in /app/migrations/*.sql; do
    if [ -f "$f" ]; then
      psql -U survey_user -d survey_app -f "$f"
    fi
  done
fi

# Start Go application
cd /app
go run main.go handlers.go
