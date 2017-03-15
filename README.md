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
| :zap: zap | 1466 ns/op | 705 B/op | 2 allocs/op |
| :zap: zap (sugared) | 2893 ns/op | 1931 B/op | 21 allocs/op |
| go-kit | 8183 ns/op | 3119 B/op | 65 allocs/op |
| lion | 12259 ns/op | 5999 B/op | 62 allocs/op |
| logrus | 12862 ns/op | 5783 B/op | 77 allocs/op |
| apex/log | 20317 ns/op | 4024 B/op | 64 allocs/op |
| log15 | 31855 ns/op | 5536 B/op | 91 allocs/op |

Log a message with a logger that already has 10 fields of context:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 536 ns/op | 0 B/op | 0 allocs/op |
| :zap: zap (sugared) | 734 ns/op | 80 B/op | 2 allocs/op |
| lion | 6784 ns/op | 3978 B/op | 36 allocs/op |
| go-kit | 8316 ns/op | 2950 B/op | 50 allocs/op |
| logrus | 10160 ns/op | 3967 B/op | 61 allocs/op |
| apex/log | 17095 ns/op | 2801 B/op | 49 allocs/op |
| log15 | 19112 ns/op | 2545 B/op | 42 allocs/op |

Log a static string, without any context or `printf`-style templating:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 521 ns/op | 0 B/op | 0 allocs/op |
| standard library | 580 ns/op | 80 B/op | 2 allocs/op |
| :zap: zap (sugared) | 885 ns/op | 80 B/op | 2 allocs/op |
| go-kit | 1384 ns/op | 656 B/op | 13 allocs/op |
| lion | 2009 ns/op | 1224 B/op | 10 allocs/op |
| logrus | 2925 ns/op | 1409 B/op | 25 allocs/op |
| apex/log | 3723 ns/op | 584 B/op | 11 allocs/op |
| log15 | 6349 ns/op | 1496 B/op | 24 allocs/op |

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
[cov-img]: https://coveralls.io/repos/github/uber-go/zap/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber-go/zap?branch=master
[benchmarking suite]: https://github.com/uber-go/zap/tree/master/benchmarks
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
