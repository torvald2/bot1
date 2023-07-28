APP?=treasury_service
RELEASE?=${shell git describe --tags $(git rev-list --tags --max-count=1)}
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
PROJECT?=treasury_service
GOOS?=linux
GOARCH?=amd64

clean:
	rm -f ${APP}

build: clean
	GOOS=${GOOS} GOARCH=${GOARCH} go build \
	-o ${APP}