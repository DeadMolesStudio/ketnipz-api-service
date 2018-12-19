FROM golang:alpine as builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -mod vendor -a -installsuffix cgo -ldflags="-w -s" -o dmstudio-server

FROM alpine

WORKDIR /app
COPY --from=builder /src/dmstudio-server .
COPY logger/logger-config.json logger/logger-config.json
COPY migrations migrations

VOLUME ["/var/log/dmstudio", "/app/static"]

ENV db_connstr ${db_connstr}
ENV db_name ${db_name}
ENV auth_connstr ${auth_connstr}

EXPOSE 8080
CMD ["sh", "-c", "./dmstudio-server -db_connstr ${db_connstr} -db_name ${db_name} -auth_connstr ${auth_connstr}"]
