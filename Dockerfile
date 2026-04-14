FROM golang:1.26.2 AS build

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/shadowmere-helper-bot /usr/bin/

ENTRYPOINT [ "/usr/bin/shadowmere-helper-bot" ]
