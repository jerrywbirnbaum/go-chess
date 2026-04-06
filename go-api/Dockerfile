FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go .

RUN CGO_ENABLED=0 go build -o /go-chess

FROM scratch 

COPY --from=builder /go-chess /go-chess

EXPOSE 8080

ENV GIN_MODE release

CMD ["/go-chess"]
