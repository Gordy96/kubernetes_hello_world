# syntax=docker/dockerfile:1
FROM golang:1.16 AS builder
ADD . /go/src/goinv
WORKDIR /go/src/goinv/
RUN go get -v all  
RUN go mod vendor  
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/server.go

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/goinv/server .
EXPOSE 8000
CMD ["./server"]