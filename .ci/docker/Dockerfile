FROM golang:1.24-alpine as builder

RUN apk update && apk add git

WORKDIR /app
COPY . .

ENV CGO_ENABLED 0
ENV GO111MODULE=on

RUN go build -o app ./.

FROM alpine:3.18

RUN apk update && apk add curl

WORKDIR /app

COPY --from=builder /app/app /app/app

EXPOSE 80

ENTRYPOINT ["/app/app"]
