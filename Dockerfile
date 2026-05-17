FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /addlicense ./cmd/addlicense

FROM scratch
COPY --from=builder /addlicense /addlicense
ENTRYPOINT ["/addlicense"]
