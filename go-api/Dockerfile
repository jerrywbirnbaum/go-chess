FROM golang:1.26
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-chess
EXPOSE 8080

# Run
CMD ["/go-chess"]
