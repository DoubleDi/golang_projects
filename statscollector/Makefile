PKGS = $(shell go list ./... | grep -v /test)

vet:
	go vet $(PKGS) 
.PHONY: vet

test-unit:
	go test --race --cover -v $(PKGS)
.PHONY: test-unit

test-benchmark:
	go test -v -bench=. statscollector.go statscollector_test.go -benchtime=20s
.PHONY: test-benchmark

test: lint test-unit test-benchmark
.PHONY: test
