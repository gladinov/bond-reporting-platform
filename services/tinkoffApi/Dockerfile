FROM golang:1.25.1-alpine AS builder

WORKDIR /usr/local/src

RUN apk add --no-cache bash git make gettext

# dependencies
COPY ["go.mod","go.sum","./"]
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/tinkoffApi ./cmd/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/tinkoffApi /

CMD ["/tinkoffApi"]
