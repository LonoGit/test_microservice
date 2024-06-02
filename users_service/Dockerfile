FROM golang:1.20.4
RUN mkdir /app
ADD . /app/
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o users_service .
CMD ["/app/users_service"]