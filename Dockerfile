FROM golang:1.18 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY types.go types.go
COPY exporter.go exporter.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o hetzner-inventory-exporter . ; strip hetzner-inventory-exporter

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/hetzner-inventory-exporter /
USER 65532:65532

EXPOSE 9112

ENTRYPOINT ["/hetzner-inventory-exporter"]
