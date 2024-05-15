.PHONY: run
run:
	go run cmd/main.go

.PHONY: test
test:
	go test ./...

.PHONY: run-bitcoind
run-bitcoind:
	./bitcoin/src/bitcoind -regtest -debug=1
