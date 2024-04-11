###############
# base images #
###############
FROM golang:1.22-bookworm as server-base
FROM alpine:3.15.4 as final


############################
# server development stage #
############################
FROM server-base as server-devel
WORKDIR /go/src/github.com/rclsilver/monitoring/daemon
COPY . /go/src/github.com/rclsilver/monitoring/daemon
ENTRYPOINT ["go", "run", "main.go"]


########################
# build Go application #
########################
FROM server-base as server-build
WORKDIR /go/src/github.com/rclsilver/monitoring/daemon
COPY . /go/src/github.com/rclsilver/monitoring/daemon
RUN make daemon && ls -l


#####################
# build final image #
#####################
FROM final
WORKDIR /

COPY --from=server-build /go/src/github.com/rclsilver/monitoring/daemon/daemon /monitoring-daemon

EXPOSE 8080

ENTRYPOINT [ "/monitoring-daemon" ]
