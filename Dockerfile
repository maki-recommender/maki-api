# Build image
FROM golang:1.18 as builder

COPY anime anime
COPY conf conf
COPY datafetch datafetch
COPY models models
COPY protos protos
COPY main.go .
COPY go.mod .
COPY go.sum .

ENV GOPATH /

RUN go mod download
RUN go mod verify
RUN go build -o /maki

# actual image
FROM debian:bullseye-slim

COPY --from=builder /maki /maki

EXPOSE 8080
CMD ["./maki"]