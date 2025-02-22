FROM golang:1.22.9

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY internal/ internal/
COPY .env ./
COPY postgres/ postgres/

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /calenduh-api

CMD [ "/calenduh-api" ]