# :zap: zap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Blazing fast, structured, leveled logging in Go.

## Installation

`go get -u go.uber.org/zap`

Note that zap only supports the two most recent minor versions of Go.

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than than other structured logging libraries
and includes both structured and `printf`-style APIs.

```go
logger, _ := zap.NewProduction()
defer logger.Sync() // flushes buffer, if any
sugar := logger.Sugar()
sugar.Infow("Failed to fetch URL.",
  // Structured context as loosely-typed key-value pairs.
  "url", url,
  "attempt", 3,
  "backoff", time.Second,
)
sugar.Infof("Failed to fetch URL: %s", url)
```

When performance and type safety are critical, use the `Logger`. It's even faster than
the `SugaredLogger` and allocates far less, but it only supports structured logging.

```go
logger, _ := zap.NewProduction()
defer logger.Sync()
logger.Info("Failed to fetch URL.",
  // Structured context as strongly-typed Field values.
  zap.String("url", url),
  zap.Int("attempt", 3),
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
| :zap: zap | 1526 ns/op | 704 B/op | 2 allocs/op |
| :zap: zap (sugared) | 2274 ns/op | 1610 B/op | 20 allocs/op |
| go-kit | 5854 ns/op | 2895 B/op | 66 allocs/op |
| logrus | 9117 ns/op | 6092 B/op | 78 allocs/op |
| lion | 9408 ns/op | 5807 B/op | 63 allocs/op |
| apex/log | 17007 ns/op | 3832 B/op | 65 allocs/op |
| log15 | 21290 ns/op | 5632 B/op | 93 allocs/op |

Log a message with a logger that already has 10 fields of context:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 446 ns/op | 0 B/op | 0 allocs/op |
| :zap: zap (sugared) | 599 ns/op | 80 B/op | 2 allocs/op |
| lion | 5231 ns/op | 4074 B/op | 38 allocs/op |
| go-kit | 6424 ns/op | 3046 B/op | 52 allocs/op |
| logrus | 7578 ns/op | 4564 B/op | 63 allocs/op |
| apex/log | 15697 ns/op | 2898 B/op | 51 allocs/op |
| log15 | 15879 ns/op | 2642 B/op | 44 allocs/op |

Log a static string, without any context or `printf`-style templating:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 418 ns/op | 0 B/op | 0 allocs/op |
| standard library | 524 ns/op | 80 B/op | 2 allocs/op |
| :zap: zap (sugared) | 628 ns/op | 80 B/op | 2 allocs/op |
| go-kit | 1011 ns/op | 656 B/op | 13 allocs/op |
| lion | 1382 ns/op | 1224 B/op | 10 allocs/op |
| logrus | 2263 ns/op | 1505 B/op | 27 allocs/op |
| apex/log | 3198 ns/op | 584 B/op | 11 allocs/op |
| log15 | 5787 ns/op | 1592 B/op | 26 allocs/op |

## Development Status: Stable
All APIs are finalized, and no breaking changes will be made in the 1.x series
of releases. Users of semver-aware dependency management systems should pin zap
to `^1`.

<hr>

Released under the [MIT License](LICENSE.txt).

<sup id="footnote-versions">1</sup> In particular, keep in mind that we may be
benchmarking against slightly older versions of other libraries. Versions are
pinned in zap's [glide.lock][] file. [↩](#anchor-versions)

[doc-img]: https://godoc.org/go.uber.org/zap?status.svg
[doc]: https://godoc.org/go.uber.org/zap
[ci-img]: https://travis-ci.org/uber-go/zap.svg?branch=master
[ci]: https://travis-ci.org/uber-go/zap
[cov-img]: https://codecov.io/gh/uber-go/zap/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/uber-go/zap
[benchmarking suite]: https://github.com/uber-go/zap/tree/master/benchmarks
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
