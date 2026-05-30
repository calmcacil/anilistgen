FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /usr/local/bin/animelistgen -ldflags='-s -w' ./cmd/animelistgen

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /usr/local/bin/animelistgen /usr/local/bin/animelistgen

ENTRYPOINT ["animelistgen"]
CMD ["daemon"]
