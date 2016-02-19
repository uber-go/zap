# Copyright (c) 2016 Uber Technologies, Inc.

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

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
