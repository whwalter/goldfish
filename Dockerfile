FROM golang:1.16 as build

COPY . /go/src/goldfish
WORKDIR /go/src/goldfish
RUN go build -o /go/bin/goldfish main.go

FROM debian:buster-slim
RUN groupadd -r goldie && useradd --no-log-init -r -g goldie goldie
COPY --from=build  /go/bin/goldfish /bin/goldfish
RUN chmod +x /bin/goldfish
USER goldie
ENTRYPOINT ["/bin/goldfish"]
