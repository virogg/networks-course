FROM golang:1.26 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service ./cmd/store

FROM gcr.io/distroless/base-debian12 as service
WORKDIR /
COPY --from=builder /app/service /store-service
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/store-service"]