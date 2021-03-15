FROM golang:latest as builder
WORKDIR /app/
COPY go.mod go.sum /app/
COPY migrations/ /app/migrations/
RUN go mod download
COPY . /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/rsscollector cmd/rsscollector/main.go

FROM debian:stable-slim
COPY --from=builder /app/rsscollector /.
COPY --from=builder /app/migrations /migrations
EXPOSE 8080
CMD /rsscollector
