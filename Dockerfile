FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download && go mod verify && go mod tidy

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN --mount=type=cache,target="/root/.cache/go-build" go build -o wolserver ./cmd/server


FROM alpine:3.18 AS runtime

WORKDIR /app

# Copy only the binary and the .env file
COPY --from=builder /app/wolserver /app/wolserver
COPY .env /app/.env

RUN chmod +x /app/wolserver

CMD ["/app/wolserver"]
