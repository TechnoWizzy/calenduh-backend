FROM golang:1.22.9

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY internal/ internal/
COPY .env ./
COPY postgres/ postgres/
COPY sqlc.yml ./

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN sqlc generate
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /calenduh-api

CMD [ "/calenduh-api" ]