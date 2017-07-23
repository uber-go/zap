# :zap: zap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Blazing fast, structured, leveled logging in Go.

## Installation

`go get -u go.uber.org/zap`

Note that zap only supports the two most recent minor versions of Go.

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than than other structured logging
packages and includes both structured and `printf`-style APIs.

```go
logger, _ := zap.NewProduction()
defer logger.Sync() // flushes buffer, if any
sugar := logger.Sugar()
sugar.Infow("failed to fetch URL",
  // Structured context as loosely typed key-value pairs.
  "url", url,
  "attempt", 3,
  "backoff", time.Second,
)
sugar.Infof("Failed to fetch URL: %s", url)
```

When performance and type safety are critical, use the `Logger`. It's even
faster than the `SugaredLogger` and allocates far less, but it only supports
structured logging.

```go
logger, _ := zap.NewProduction()
defer logger.Sync()
logger.Info("failed to fetch URL",
  // Structured context as strongly typed Field values.
  zap.String("url", url),
  zap.Int("attempt", 3),
  zap.Duration("backoff", time.Second),
)
```

See the [documentation][doc] and [FAQ](FAQ.md) for more details.

## Performance

For applications that log in the hot path, reflection-based serialization and
string formatting are prohibitively expensive &mdash; they're CPU-intensive
and make many small allocations. Put differently, using `encoding/json` and
`fmt.Fprintf` to log tons of `interface{}`s makes your application slow.

Zap takes a different approach. It includes a reflection-free, zero-allocation
JSON encoder, and the base `Logger` strives to avoid serialization overhead
and allocations wherever possible. By building the high-level `SugaredLogger`
on that foundation, zap lets users *choose* when they need to count every
allocation and when they'd prefer a more familiar, loosely typed API.

As measured by its own [benchmarking suite][], not only is zap more performant
than comparable structured logging packages &mdash; it's also faster than the
standard library. Like all benchmarks, take these with a grain of salt.<sup
id="anchor-versions">[1](#footnote-versions)</sup>

Log a message and 10 fields:

| Package | Time | Objects Allocated |
| :--- | :---: | :---: |
| :zap: zap | 3174 ns/op | 5 allocs/op |
| :zap: zap (sugared) | 3929 ns/op | 21 allocs/op |
| go-kit | 16475 ns/op | 126 allocs/op |
| zerolog | 17013 ns/op | 90 allocs/op |
| lion | 17119 ns/op | 111 allocs/op |
| logrus | 23344 ns/op | 142 allocs/op |
| log15 | 34996 ns/op | 149 allocs/op |
| apex/log | 39062 ns/op | 126 allocs/op |

Log a message with a logger that already has 10 fields of context:

| Package | Time | Objects Allocated |
| :--- | :---: | :---: |
| :zap: zap | 381 ns/op | 0 allocs/op |
| :zap: zap (sugared) | 546 ns/op | 2 allocs/op |
| zerolog | 316 ns/op | 0 allocs/op |
| lion | 6215 ns/op | 39 allocs/op |
| go-kit | 19241 ns/op | 115 allocs/op |
| logrus | 21331 ns/op | 130 allocs/op |
| log15 | 26071 ns/op | 79 allocs/op |
| apex/log | 37891 ns/op | 115 allocs/op |

Log a static string, without any context or `printf`-style templating:

| Package | Time | Objects Allocated |
| :--- | :---: | :---: |
| :zap: zap | 338 ns/op | 0 allocs/op |
| :zap: zap (sugared) | 549 ns/op | 2 allocs/op |
| zerolog | 303 ns/op | 0 allocs/op |
| standard library | 586 ns/op | 2 allocs/op |
| go-kit | 958 ns/op | 13 allocs/op |
| lion | 1452 ns/op | 10 allocs/op |
| logrus | 2367 ns/op | 27 allocs/op |
| apex/log | 3383 ns/op | 11 allocs/op |
| log15 | 6481 ns/op | 26 allocs/op |

## Development Status: Stable

All APIs are finalized, and no breaking changes will be made in the 1.x series
of releases. Users of semver-aware dependency management systems should pin
zap to `^1`.

<hr>

Released under the [MIT License](LICENSE.txt).

<sup id="footnote-versions">1</sup> In particular, keep in mind that we may be
benchmarking against slightly older versions of other packages. Versions are
pinned in zap's [glide.lock][] file. [↩](#anchor-versions)

[doc-img]: https://godoc.org/go.uber.org/zap?status.svg
[doc]: https://godoc.org/go.uber.org/zap
[ci-img]: https://travis-ci.org/uber-go/zap.svg?branch=master
[ci]: https://travis-ci.org/uber-go/zap
[cov-img]: https://codecov.io/gh/uber-go/zap/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/uber-go/zap
[benchmarking suite]: https://github.com/uber-go/zap/tree/master/benchmarks
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
