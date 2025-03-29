# Step 1: Build the migrate tool with SQLite3 support
FROM golang:1.20-alpine as builder

RUN apk add --no-cache build-base

WORKDIR /app

# Install the migrate tool with SQLite3 driver support
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Step 2: Copy migration files into the container
COPY ./db/migrations /migrations

# Step 3: Run the migrations
CMD ["migrate", "-database", "sqlite3://database.db", "-path", "/migrations", "up"]
