.PHONY: test build demo-ab oracle-check

test:
	go test ./...

build:
	go build ./cmd/gossip ./cmd/gossipctl

demo-ab:
	sh scripts/demo-ab-build.sh

oracle-check:
	sh scripts/check-oracle-fixture.sh
