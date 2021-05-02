# Build stage
FROM golang:1.16.3 AS build-env
ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR /app/certm
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN go build -o certm
RUN chmod +x certm

# Release stage
FROM alpine:latest AS release
WORKDIR /app/
COPY --from=build-env /app/certm/certm .
ENV WORKDIR "/app/"
ENV PATH "${WORKDIR}:${PATH}"

CMD ["certm", "--version"]
