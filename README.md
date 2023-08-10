# :zap: zap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Blazing fast, structured, leveled logging in Go.

## Installation

`go get -u go.uber.org/zap`

Note that zap only supports the two most recent minor versions of Go.

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than other structured logging
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

| Package | Time | Time % to zap | Objects Allocated |
| :------ | :--: | :-----------: | :---------------: |
| :zap: zap | 645 ns/op | +0% | 5 allocs/op
| :zap: zap (sugared) | 978 ns/op | +52% | 11 allocs/op
| zerolog | 237 ns/op | -63% | 1 allocs/op
| slog | 2215 ns/op | +243% | 40 allocs/op
| go-kit | 2389 ns/op | +270% | 58 allocs/op
| fortio log | 4798 ns/op | +644% | 120 allocs/op
| apex/log | 9909 ns/op | +1436% | 64 allocs/op
| log15 | 11405 ns/op | +1668% | 75 allocs/op
| logrus | 11966 ns/op | +1755% | 80 allocs/op

Log a message with a logger that already has 10 fields of context:

| Package | Time | Time % to zap | Objects Allocated |
| :------ | :--: | :-----------: | :---------------: |
| :zap: zap | 63 ns/op | +0% | 0 allocs/op
| :zap: zap (sugared) | 80 ns/op | +27% | 1 allocs/op
| zerolog | 26 ns/op | -59% | 0 allocs/op
| slog | 244 ns/op | +287% | 0 allocs/op
| go-kit | 2517 ns/op | +3895% | 56 allocs/op
| fortio log | 4430 ns/op | +6932% | 112 allocs/op
| log15 | 9307 ns/op | +14673% | 70 allocs/op
| apex/log | 9341 ns/op | +14727% | 53 allocs/op
| logrus | 10952 ns/op | +17284% | 68 allocs/op

Log a static string, without any context or `printf`-style templating:

| Package | Time | Time % to zap | Objects Allocated |
| :------ | :--: | :-----------: | :---------------: |
| :zap: zap | 62 ns/op | +0% | 0 allocs/op
| :zap: zap (sugared) | 78 ns/op | +26% | 1 allocs/op
| zerolog | 23 ns/op | -63% | 0 allocs/op
| slog | 233 ns/op | +276% | 0 allocs/op
| go-kit | 247 ns/op | +298% | 9 allocs/op
| standard library | 293 ns/op | +373% | 2 allocs/op
| fortio log | 430 ns/op | +594% | 8 allocs/op
| apex/log | 786 ns/op | +1168% | 5 allocs/op
| logrus | 1477 ns/op | +2282% | 23 allocs/op
| log15 | 2029 ns/op | +3173% | 20 allocs/op

## Development Status: Stable

All APIs are finalized, and no breaking changes will be made in the 1.x series
of releases. Users of semver-aware dependency management systems should pin
zap to `^1`.

## Contributing

We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The zap maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
oss-conduct@uber.com. That email list is a private, safe space; even the zap
maintainers don't have access, so don't hesitate to hold us to a high
standard.

<hr>

Released under the [MIT License](LICENSE.txt).

<sup id="footnote-versions">1</sup> In particular, keep in mind that we may be
benchmarking against slightly older versions of other packages. Versions are
pinned in the [benchmarks/go.mod][] file. [↩](#anchor-versions)

[doc-img]: https://pkg.go.dev/badge/go.uber.org/zap
[doc]: https://pkg.go.dev/go.uber.org/zap
[ci-img]: https://github.com/uber-go/zap/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/uber-go/zap/actions/workflows/go.yml
[cov-img]: https://codecov.io/gh/uber-go/zap/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/uber-go/zap
[benchmarking suite]: https://github.com/uber-go/zap/tree/master/benchmarks
[benchmarks/go.mod]: https://github.com/uber-go/zap/blob/master/benchmarks/go.mod

