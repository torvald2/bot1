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
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
	-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
	-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
	-o ${APP}