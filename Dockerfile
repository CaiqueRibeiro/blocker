FROM golang:latest

WORKDIR /app

RUN apt-get update
RUN apt-get install -y protobuf-compiler

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

RUN export PATH="$PATH:$(go env GOPATH)/bin"

CMD ["tail", "-f", "/dev/null"]