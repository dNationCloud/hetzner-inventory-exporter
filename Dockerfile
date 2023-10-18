FROM golang:1.20 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o exporter . ; strip exporter

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/exporter /
USER 65532:65532

EXPOSE 9112

ENTRYPOINT ["/exporter"]
