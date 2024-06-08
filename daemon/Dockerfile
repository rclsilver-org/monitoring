###############
# base images #
###############
FROM golang:1.22-alpine as server-base
FROM scratch as final


############################
# server development stage #
############################
FROM server-base as server-devel
COPY . /go/src/github.com/rclsilver/monitoring/daemon
WORKDIR /go/src/github.com/rclsilver/monitoring/daemon
ENTRYPOINT ["go", "run", "main.go"]


########################
# build Go application #
########################
FROM server-base as server-build
COPY . /go/src/github.com/rclsilver/monitoring
WORKDIR /go/src/github.com/rclsilver/monitoring/daemon
RUN apk add --no-cache make && \
    make dist/daemon-amd64


#####################
# build final image #
#####################
FROM final
WORKDIR /

COPY --from=server-build /go/src/github.com/rclsilver/monitoring/daemon/dist/daemon-amd64 /monitoring-daemon

EXPOSE 8080

ENTRYPOINT [ "/monitoring-daemon" ]
