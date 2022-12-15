FROM golang:1.19 as build

COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.15
COPY --from=build /app/shadowmere-helper-bot /user/bin/

ENTRYPOINT /user/bin/shadowmere-helper-bot
