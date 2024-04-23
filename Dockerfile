FROM golang:1.16 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
# RUN ls
RUN CGO_ENABLED=0 GOOS=linux go build -o /main

FROM scratch
WORKDIR /
COPY --from=builder /main /main
EXPOSE 8080
ENTRYPOINT ["./main"]
