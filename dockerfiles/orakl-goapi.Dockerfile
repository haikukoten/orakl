
FROM golang:1.21.5-bullseye as builder

RUN apt-get update && apt-get install -y curl

WORKDIR /app

COPY go-api go-api

WORKDIR /app/go-api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o apibin -ldflags="-w -s" .

# debian:bullseye-slim
FROM debian@sha256:4b48997afc712259da850373fdbc60315316ee72213a4e77fc5a66032d790b2a

RUN apt-get update && apt-get install -y curl

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv ./migrate /usr/bin

WORKDIR /app

RUN mkdir /app/migrations

COPY --from=builder /app/go-api/migrations /app/migrations

COPY --from=builder /app/go-api/apibin /usr/bin

CMD sh -c '_DATABASE_URL=$DATABASE_URL; if echo $_DATABASE_URL | grep -q "\?"; then \
              _DATABASE_URL="${_DATABASE_URL}&sslmode=disable"; \
           else \
              _DATABASE_URL="${_DATABASE_URL}?sslmode=disable"; \
           fi && \
           migrate -database "$_DATABASE_URL" -verbose -path ./migrations up && apibin'