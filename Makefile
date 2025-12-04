export GO111MODULE = on

.PHONY: default test test-cover bench lint


test:
	go test -race -cover ./...

test-cover:
	go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

bench:
	go test --benchmem -benchtime=10s -bench='Benchmark.*' -run='^$$'

lint:
	golangci-lint run --timeout=600s && go vet ./...

