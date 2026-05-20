FROM golang:1.25.4

WORKDIR /server
COPY . .

RUN go mod tidy
RUN go build ./cmd/main.go

EXPOSE 8080

CMD [ "./main" ]