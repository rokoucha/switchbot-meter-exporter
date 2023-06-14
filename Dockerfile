FROM docker.io/library/golang:1.20-bullseye AS build

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download

COPY *.go /app/

RUN go build -o /app/swichbot-meter-exporter

FROM docker.io/library/debian:bullseye-slim

ARG GROUP_ID="998"
ARG USER_NAME="swichbot-meter-exporter"
ARG USER_ID="998"

RUN groupadd -g "${GROUP_ID}" "${USER_NAME}" && \
    useradd -l -u "${USER_ID}" -g "${USER_NAME}" "${USER_NAME}"

WORKDIR /app

COPY --from=build /app/swichbot-meter-exporter /app/swichbot-meter-exporter

USER $USER_NAME

ENTRYPOINT ["/app/swichbot-meter-exporter"]
