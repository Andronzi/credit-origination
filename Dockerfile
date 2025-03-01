FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# RUN apk add --no-cache git make musl-dev gcc

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]

# RUN apk add --no-cache bash netcat-openbsd
# RUN apk add --no-cache bash ca-certificates

# COPY --from=builder /app/docker/wait-for-it.sh .

# RUN chmod +x wait-for-it.sh

# CMD ["./wait-for-it.sh", "db:5432", "--", "./main"]