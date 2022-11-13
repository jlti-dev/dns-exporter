FROM golang:1.17 as builder
WORKDIR /app

RUN go get github.com/miekg/dns && \
    go get github.com/prometheus/client_golang/prometheus && \
    go get github.com/prometheus/client_golang/prometheus/promauto && \
    go get github.com/prometheus/client_golang/prometheus/promhttp

COPY app/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin .
CMD ["/app/main"]
HEALTHCHECK --interval=30s --timeout=1s CMD wget --spider -O /dev/null localhost:8080/metrics
