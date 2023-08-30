FROM golang:1.21 AS build

COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.18
COPY --from=build /app/shadowmere-helper-bot /user/bin/

ENTRYPOINT /user/bin/shadowmere-helper-bot
