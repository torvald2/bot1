FROM golang:1.19

WORKDIR /dockerapp

COPY . .

RUN go mod download

RUN go build -o /main


CMD ["/main"]