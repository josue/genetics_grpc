FROM golang:1.12-alpine

RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN go get -d ./...
RUN go build -o /app/bin .

EXPOSE 50051
CMD ["/app/bin"]