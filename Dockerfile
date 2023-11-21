FROM golang:1.21.4-alpine3.18
LABEL authors="artemy"

WORKDIR /app

COPY go.mod /app
COPY go.sum /app
RUN go mod download

COPY . /app

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-golang-app notifications-bot/cmd

EXPOSE 55000

# Run
CMD ["/docker-golang-app"]

