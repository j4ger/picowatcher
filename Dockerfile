FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/picowatcher .

FROM alpine:3.21

RUN addgroup -S picowatcher && adduser -S -G picowatcher picowatcher

WORKDIR /data
COPY --from=builder /out/picowatcher /usr/local/bin/picowatcher

USER picowatcher

ENTRYPOINT ["/usr/local/bin/picowatcher"]
