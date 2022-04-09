FROM golang:1.17 AS build

WORKDIR /go/src/container-monitor

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/server.go

FROM alpine:3.9

COPY --from=build /go/src/container-monitor/bin/server .

CMD ["./server"]
