FROM golang:alpine3.23

WORKDIR /app

COPY . . 
RUN go mod tidy
RUN go build -o restaurant-management
EXPOSE 8080

CMD ["./restaurant-management"]