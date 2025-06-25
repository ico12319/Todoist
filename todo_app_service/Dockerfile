ARG GOLANG_VERSION=1.24

FROM golang:${GOLANG_VERSION} AS builder

WORKDIR /app

COPY go.mod /app/

COPY go.sum /app/

COPY . /app/

RUN go mod download

RUN CGO_ENABLED=0 go build -o /main ./todo_app_service/cmd

ARG ALPINE_VERSION=3.20

FROM alpine:${ALPINE_VERSION}

WORKDIR /api

COPY --from=builder /main  .

COPY --from=builder /app/.env  .

EXPOSE 3434

ENTRYPOINT ["./main"]

