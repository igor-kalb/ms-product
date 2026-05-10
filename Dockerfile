FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /ms-product

FROM alpine:3.21

COPY --from=builder /ms-product /ms-product

EXPOSE 8000

CMD [ "/ms-product" ]
