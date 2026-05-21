FROM golang:1.25.1-alpine AS builder

WORKDIR /usr/local/src

RUN apk add --no-cache bash git make gettext gcc musl-dev 

# dependencies
COPY ["go.mod","go.sum","./"]
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o ./bin/gateway ./cmd/main.go

FROM alpine AS runner
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /usr/local/src/bin/gateway /

CMD ["/gateway"]