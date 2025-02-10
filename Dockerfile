FROM golang:1.23-alpine AS builder

RUN apk add protobuf-dev curl make
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

WORKDIR /app
COPY . .

RUN protoc --go_out=./api/proto/gen --go_opt=paths=source_relative \
            --go-grpc_out=./api/proto/gen --go-grpc_opt=paths=source_relative \
            --proto_path=./api/proto \
            ./api/proto/*.proto

RUN go build -o /app/server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/server /server

ENV PORT=50051
ENV STORAGE_TYPE=memory
ENV POSTGRES_URL=""

CMD ["/server"]