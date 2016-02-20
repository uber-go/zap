
BENCH_FLAGS ?= -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem
PACKAGES ?= $(shell glide novendor)

.PHONY: all
all: lint test

.PHONY: dependencies
dependencies:
	glide --version || go get -u -f github.com/Masterminds/glide
	glide install
	go install ./vendor/github.com/golang/lint/golint
	go install ./vendor/github.com/axw/gocov/gocov
	go install ./vendor/github.com/mattn/goveralls

.PHONY: lint
lint:
	rm -rf lint.log
	@echo "Checking formatting..."
	gofmt -d -s *.go benchmarks 2>&1 | tee lint.log
	@echo "Checking vet..."
	go tool vet *.go 2>&1 | tee -a lint.log
	go tool vet benchmarks 2>&1 | tee -a lint.log
	@echo "Checking lint..."
	golint . 2>&1 | tee -a lint.log
	golint benchmarks 2>&1 | tee -a lint.log
	@[ ! -s lint.log ]

.PHONY: test
test:
	go test -race $(PACKAGES)

.PHONY: coveralls
coveralls:
	goveralls -service=travis-ci $(PACKAGES)

.PHONY: bench
BENCH ?= .
bench:
	go test -bench=$(BENCH) $(BENCH_FLAGS) ./benchmarks
