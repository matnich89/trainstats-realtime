FROM golang:1.23 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/main.go

EXPOSE 8080

FROM scratch AS production
COPY --from=builder /app .
CMD ["./app"]