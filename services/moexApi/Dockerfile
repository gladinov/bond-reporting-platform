FROM golang:1.25.1-alpine AS builder

WORKDIR /usr/local/src

RUN apk add --no-cache bash git make gettext

# dependencies
COPY ["go.mod","go.sum","./"]
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/moexApi ./cmd/main.go

FROM alpine AS runner
RUN apk add --no-cache tzdata

COPY --from=builder /usr/local/src/bin/moexApi /

CMD ["/moexApi"]
