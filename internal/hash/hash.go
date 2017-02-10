// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap/internal/hash"
)

var _n uint32

var hashes = map[string]func(k string) uint32{
	"xshrr":  hash.XSHRR,
	"fnv32":  hash.FNV32,
	"fnv32a": hash.FNV32a,
}

func main() {
	hn := flag.String("hash", "xshrr", "hash function")
	n := flag.Uint("n", 4096, "mod-N")
	flag.Parse()

	h := hashes[*hn]
	if h == nil {
		log.Fatalf("invalid hash function %q", *hn)
	}

	_n = uint32(*n)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		fmt.Println(h(sc.Text()) % _n)
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}
