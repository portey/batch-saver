FROM golang:1.14.6-alpine as builder
WORKDIR /app
COPY . ./
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -o ./bin/svc

FROM scratch
COPY --from=builder /app/bin/svc /svc
COPY --from=builder /app/storage/postgres/migrations /storage/postgres/migrations
EXPOSE 8080 8888
CMD ["./svc"]
