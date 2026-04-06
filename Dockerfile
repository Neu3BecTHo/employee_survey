FROM golang:1.24-alpine

WORKDIR /app

# Install PostgreSQL, git and su-exec
RUN apk add --no-cache postgresql postgresql-contrib git su-exec

# Initialize PostgreSQL
RUN mkdir -p /var/lib/postgresql/data /run/postgresql && \
    chown -R postgres:postgres /var/lib/postgresql /run/postgresql && \
    su - postgres -c "initdb -D /var/lib/postgresql/data"

# Configure PostgreSQL
RUN echo "host all all 127.0.0.1/32 trust" >> /var/lib/postgresql/data/pg_hba.conf && \
    echo "listen_addresses='127.0.0.1'" >> /var/lib/postgresql/data/postgresql.conf

# Create database and user
RUN su - postgres -c "pg_ctl start -D /var/lib/postgresql/data" && \
    sleep 2 && \
    su - postgres -c "psql -c \"CREATE USER survey_user WITH PASSWORD 'survey_pass';\"" && \
    su - postgres -c "psql -c \"CREATE DATABASE survey_app OWNER survey_user;\"" && \
    su - postgres -c "pg_ctl stop -D /var/lib/postgresql/data"

# Install Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application code
COPY . .

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
