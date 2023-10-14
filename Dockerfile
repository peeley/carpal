FROM golang:1.21 AS build

WORKDIR /app

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /carpal cmd/main.go

FROM alpine:latest

EXPOSE 8008

COPY ./configs/default.yml /etc/carpal/config.yml
COPY --from=build /carpal /

CMD ["/carpal"]
