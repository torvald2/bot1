FROM golang:1.19 AS build_base
WORKDIR /main
COPY . /main
RUN apk update
RUN apk add git
RUN apk add build-essential

RUN apk add make

RUN make build

# Start fresh from a smaller image
FROM alpine:3.9
COPY --from=build_base /main/treasury_service /
COPY . .
CMD ["/treasury_service"]