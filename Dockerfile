FROM golang:alpine as BUILD
RUN apk update && apk add curl git make bash openssh-client && rm -rf /var/lib/apk/cache/*
WORKDIR /go/src/microplane
COPY . .
RUN make install_deps && make build

FROM alpine/git
COPY --from=BUILD /go/src/microplane/bin/mp /bin/mp
ENTRYPOINT ["/bin/mp"]
