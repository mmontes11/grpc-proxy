# grpc-proxy
HTTP/2 reverse proxy library for routing to gRPC microservices

### Usage

```
$ GODEBUG=http2debug=2 go run main.go -port 8000 -server_addr localhost:11000
```
