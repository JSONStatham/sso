FROM golang:1.24-alpine

# Install curl
RUN apk add --no-cache curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /app/bin/sso ./cmd/sso/main.go


EXPOSE 80

CMD ["/app/bin/sso"]