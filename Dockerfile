FROM docker.io/golang:1.25.1-alpine AS builder
RUN apk add make protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.8
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
COPY --from=docker.io/bufbuild/buf:1.57.0 /usr/local/bin/buf /usr/local/bin/buf
COPY . .
RUN CGO_ENABLED=0 make

FROM scratch
COPY --from=builder /go/out/nikola-telemetry /nikola-telemetry
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/nikola-telemetry"]
