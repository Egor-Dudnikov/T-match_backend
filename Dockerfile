# docker run -p 8080:8080 --env-file .env t-match-app

FROM golang:1.25-alpine

WORKDIR /app

COPY . .

RUN go build -o t-match ./cmd/main.go

CMD ["./t-match"]