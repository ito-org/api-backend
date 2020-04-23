FROM golang:alpine

COPY . /api-backend
WORKDIR /api-backend

RUN GOARCH=amd64 GOOS=linux go build -o backend github.com/ito-org/go-backend

FROM golang:alpine

RUN apk add --no-cache --upgrade bash postgresql-client

COPY --from=0 /api-backend/backend /backend
COPY ./scripts/wait-for-postgres.sh /wait-for-postgres.sh

WORKDIR /

CMD ["sh", "./wait-for-postgres.sh", "postgres", "--", "./backend", "--dbhost=postgres"]