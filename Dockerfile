# build environment
FROM golang:1.15-alpine as build-env

# Setup
RUN apk add --update --no-cache dumb-init
RUN adduser --uid 1500 -D stonks

# Copy only needed packages
COPY ./vendor $GOPATH/src/github.com/markliederbach/stonks/vendor
COPY ./pkg $GOPATH/src/github.com/markliederbach/stonks/pkg
COPY ./go.mod $GOPATH/src/github.com/markliederbach/stonks/go.mod
COPY ./go.sum $GOPATH/src/github.com/markliederbach/stonks/go.sum

# Build
WORKDIR $GOPATH/src/github.com/markliederbach/stonks/
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    go build --ldflags "-extldflags '-static'" -o /src/stonks pkg/alpaca/main.go

# Build real container from scratch
FROM scratch

ENV APP=/usr/local/bin/stonks \
    USER_UID=1500 \
    USER_NAME=stonks

USER ${USER_UID}

COPY --from=build-env /src/stonks ${APP}
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /usr/bin/dumb-init /usr/bin/dumb-init

ENTRYPOINT ["dumb-init"]
CMD ["/usr/local/bin/stonks"]

# Default port to expose
# Override with -p 9000:<host port>/tcp
# EXPOSE 9000/tcp