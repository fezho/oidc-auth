FROM golang:1.12-alpine

# Install some dependencies needed to build the project
#RUN apk add bash ca-certificates git gcc g++ libc-dev
RUN apk add --no-cache --update alpine-sdk

WORKDIR /go/src/github.com/fezho/oidc-auth-service


# Force the go compiler to use modules and set go proxy
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn

# Get dependancies - will be cached if we won't change mod/sum
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make release-binary

FROM alpine:3.10

# OpenSSL is required so wget can query HTTPS endpoints for health checking.
RUN apk add --update ca-certificates openssl

USER 1001:1001
COPY --from=0 /go/bin/auth-service /usr/local/bin/auth-service

# Import frontend assets and set the correct CWD directory so the assets
# are in the default path.
#COPY web /web
WORKDIR /

ENTRYPOINT ["auth-service"]
