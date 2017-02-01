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

// xorstring converts a string into a uint64 by xoring together its
// codepoints. It works by accumulating into a 64-bit "ring" which gets
// rotated by the apparent "byte width" of each codepoint.
func xorstring(s string) uint64 {
	var n uint64
	for i := 0; i < len(s); i++ {
		n = ((n & 0xff) >> 56) | (n << 8)
		n ^= uint64(s[i])
	}
	return n
}

// xshrr computes a "randomly" rotated xorshift; this is the "XSH RR"
// transformation borrowed from the PCG famiily of random generators. It
// returns a 32-bit output from a 64-bit state.
func xshrr64(n uint64) uint32 {
	xorshifted := uint32(((n >> 18) ^ n) >> 27)
	rot := uint32(n >> 59)
	return (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
}

// XSHRR hashes a string using the XSH-RR construction from the PCG family of
// rando number generators. The returned number is under m (0 <= XSHRR(s, m) < m).
func XSHRR(key string, m uint32) uint32 {
	return xshrr64(xorstring(key)) % m
}
