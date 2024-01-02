FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o client client/client.go  && \
    go build -o server1 server1/server1.go  && \
    go build -o server2 server2/server2.go && \
    go build -o server3 server3/server3.go

