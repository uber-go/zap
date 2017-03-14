# :zap: zap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Blazing fast, structured, leveled logging in Go.

## Installation

`go get -u go.uber.org/zap`

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than than other structured logging libraries
and includes both structured and `printf`-style APIs.

```go
logger, _ := zap.NewProduction()
sugar := logger.Sugar()
sugar.Infow("Failed to fetch URL.",
  // Structured context as loosely-typed key-value pairs.
  "url", url,
  "attempt", retryNum,
  "backoff", time.Second,
)
sugar.Infof("Failed to fetch URL: %s", url)
```

When performance and type safety are critical, use the `Logger`. It's even faster than
the `SugaredLogger` and allocates far less, but it only supports structured logging.

```go
logger, _ := zap.NewProduction()
logger.Info("Failed to fetch URL.",
  // Structured context as strongly-typed Field values.
  zap.String("url", url),
  zap.Int("attempt", tryNum),
  zap.Duration("backoff", time.Second),
)
```

## Performance

For applications that log in the hot path, reflection-based serialization and
string formatting are prohibitively expensive &mdash; they're CPU-intensive and
make many small allocations. Put differently, using `encoding/json` and
`fmt.Fprintf` to log tons of `interface{}`s makes your application slow.

Zap takes a different approach. It includes a reflection-free, zero-allocation
JSON encoder, and the base `Logger` strives to avoid serialization overhead and
allocations wherever possible. By building the high-level `SugaredLogger` on
that foundation, zap lets users *choose* when they need to count every
allocation and when they'd prefer a more familiar, loosely-typed API.

As measured by its own [benchmarking suite][], not only is zap more performant
than comparable structured logging libraries &mdash; it's also faster than the
standard library. Like all benchmarks, take these with a grain of salt.<sup
id="anchor-versions">[1](#footnote-versions)</sup>

Log a message and 10 fields:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 670 ns/op | 705 B/op | 2 allocs/op |
| :zap: zap (sugared) | 1378 ns/op | 1613 B/op | 20 allocs/op |
| go-kit | 3824 ns/op | 2897 B/op | 66 allocs/op |
| lion | 5220 ns/op | 5811 B/op | 63 allocs/op |
| logrus | 5468 ns/op | 6100 B/op | 78 allocs/op |
| apex/log | 13576 ns/op | 3834 B/op | 65 allocs/op |
| log15 | 16969 ns/op | 5633 B/op | 93 allocs/op |

Log a message with a logger that already has 10 fields of context:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 247 ns/op | 0 B/op | 0 allocs/op |
| :zap: zap (sugared) | 346 ns/op | 80 B/op | 2 allocs/op |
| lion | 3361 ns/op | 4076 B/op | 38 allocs/op |
| go-kit | 3853 ns/op | 3048 B/op | 52 allocs/op |
| logrus | 5161 ns/op | 4568 B/op | 63 allocs/op |
| apex/log | 11464 ns/op | 2898 B/op | 51 allocs/op |
| log15 | 12148 ns/op | 2642 B/op | 44 allocs/op |

Log a static string, without any context or `printf`-style templating:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 252 ns/op | 0 B/op | 0 allocs/op |
| :zap: zap (sugared) | 386 ns/op | 80 B/op | 2 allocs/op |
| standard library | 611 ns/op | 80 B/op | 2 allocs/op |
| go-kit | 636 ns/op | 656 B/op | 13 allocs/op |
| lion | 912 ns/op | 1225 B/op | 10 allocs/op |
| logrus | 1707 ns/op | 1507 B/op | 27 allocs/op |
| apex/log | 2459 ns/op | 584 B/op | 11 allocs/op |
| log15 | 5087 ns/op | 1592 B/op | 26 allocs/op |

## Development Status: Release Candidate 3
The current release is `v1.0.0-rc.3`. No further breaking changes are *planned*
unless wider use reveals critical bugs or usability issues, but users who need
absolute stability should wait for the 1.0.0 release.

<hr>
Released under the [MIT License](LICENSE.txt).

<sup id="footnote-versions">1</sup> In particular, keep in mind that we may be
benchmarking against slightly older versions of other libraries. Versions are
pinned in zap's [glide.lock][] file. [↩](#anchor-versions)

[doc-img]: https://godoc.org/go.uber.org/zap?status.svg
[doc]: https://godoc.org/go.uber.org/zap
[ci-img]: https://travis-ci.org/uber-go/zap.svg?branch=master
[ci]: https://travis-ci.org/uber-go/zap
[cov-img]: https://coveralls.io/repos/github/uber-go/zap/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber-go/zap?branch=master
[benchmarking suite]: https://github.com/uber-go/zap/tree/master/benchmarks
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
