FROM golang:1.21.2-alpine

WORKDIR /app
COPY . ./
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/report-generator ./cmd/main.go

EXPOSE 8080

ENTRYPOINT [ "./bin/report-generator" ]
