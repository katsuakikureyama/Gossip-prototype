FROM golang:1.22.1-bookworm AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/gossip ./cmd/gossip

FROM scratch
COPY --from=build /out/gossip /gossip
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/gossip"]
CMD ["-listen", "0.0.0.0:8080"]
