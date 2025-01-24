FROM golang:bookworm AS build
LABEL authors="minz1"

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd cmd
COPY pkg pkg
COPY internal internal

RUN GOOS=linux go build -ldflags '-linkmode external -extldflags "-static"' -o /bin/snitchbot ./cmd/snitchbot

FROM debian
RUN apt-get update
RUN apt-get -y install ca-certificates
COPY --from=build /bin/snitchbot /bin/snitchbot
CMD ["/bin/snitchbot"]
