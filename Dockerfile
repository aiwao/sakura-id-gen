FROM golang:1.25.5-alpine3.23 AS builder

WORKDIR /app

COPY . .

RUN apk add --no-cache git sqlite-dev gcc musl-dev
RUN go mod download
RUN go build -o sakura_id_gen

FROM alpine
RUN apk add --no-cache sqlite-libs
WORKDIR /app

COPY --from=builder /app/sakura_id_gen ./sakura_id_gen

CMD ["./sakura_id_gen"]
