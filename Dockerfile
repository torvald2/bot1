FROM golang:1.19-alpine AS build_base
WORKDIR /main
COPY . /main
RUN apk update
RUN apk add git
RUN apk add make

RUN make build

# Start fresh from a smaller image
FROM alpine:3.9
COPY --from=build_base /main/treasury_service /
COPY mail mail
COPY templates templates
COPY static static

CMD ["/treasury_service"]