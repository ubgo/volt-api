# Must be >= the highest `go` directive across modules/*/go.mod (1.26.4).
FROM golang:1.26.4-alpine3.22

RUN apk update && apk add --no-cache git ca-certificates curl socat supervisor
RUN mkdir -p /var/log/supervisor


RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

RUN go install github.com/air-verse/air@latest


ENV CGO_ENABLED=0 GO111MODULE=on GOOS=linux 
# ENV GOFLAGS=-mod=vendor

WORKDIR /app

ARG PORT
# add it as setting option while create container out of image
EXPOSE $PORT

COPY go.work ./
COPY go.work.sum ./
COPY modules/config/go.mod ./modules/config/
COPY modules/dbent/go.mod ./modules/dbent/
COPY modules/apidash/go.mod ./modules/apidash/
COPY modules/lace/go.mod ./modules/lace/
COPY modules/scripts/go.mod ./modules/scripts/
COPY modules/examples/go.mod ./modules/examples/
COPY modules/core/go.mod ./modules/core/
COPY modules/saas/go.mod ./modules/saas/

## CUSTOM
## AI: can be modified to include other custom modules
## END CUSTOM

RUN go work vendor

## CUSTOM
# this is port forward so docker can access the services or ports running on the host machine from inside the container
# Create .bashrc with port forwarding commands
# AI: can be modified to include other port forwarding commands
RUN echo 'socat TCP-LISTEN:2303,fork TCP:host.docker.internal:2303 &' >> /root/.bashrc
RUN echo 'socat TCP-LISTEN:2311,fork TCP:host.docker.internal:2311 &' >> /root/.bashrc
RUN echo 'socat TCP-LISTEN:5432,fork TCP:host.docker.internal:5432 &' >> /root/.bashrc
RUN echo 'socat TCP-LISTEN:6379,fork TCP:host.docker.internal:6379 &' >> /root/.bashrc
## END CUSTOM

# RUN source /etc/profile

# Source profile to enable port forwarding
# CMD ["sh", "-c", "source /etc/profile && exec $0 $@"]

# Supervisord configuration for multi-service container
COPY .docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
