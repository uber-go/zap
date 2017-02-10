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

package hash

// XSHRR hashes a string using the by threading each byte of it thru the PCG32
// (XSH-RR) random number generator; between each iteration of the underlying
// RNG, a byte of input is XOR'd into the state vector. It is very similar to
// FNV64a, but with a final hardening step.
func XSHRR(s string) uint32 {
	const mul = 6364136223846793005
	var n uint64
	for i := 0; i < len(s); i++ {
		n ^= uint64(s[i])
		n *= mul
	}
	xorshifted := uint32(((n >> 18) ^ n) >> 27)
	rot := uint32(n >> 59)
	res := (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
	return res
}
