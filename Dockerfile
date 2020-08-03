############################
# STEP 1 build executable binary
############################
FROM golang:1.14 as builder
WORKDIR /app
COPY . ./
RUN go build -mod=vendor -o ./bin/svc -a .

############################
# STEP 2 build a small image
############################
FROM scratch

COPY --from=builder /app/bin/svc /svc

EXPOSE 8080 8888

CMD ["./svc"]

