FROM golang:1.22.0 as builder

RUN apt-get update && apt-get install -y libpcap-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o netrics .

CMD ["./netrics"]
