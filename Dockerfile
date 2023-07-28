FROM golang:1.19 AS build_base
WORKDIR /main
COPY . /main

RUN make build

# Start fresh from a smaller image
FROM alpine:3.9
COPY --from=build_base /main/treasury_service /
COPY . .
CMD ["/treasury_service"]