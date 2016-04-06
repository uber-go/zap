# :zap: zap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Fast, structured, leveled logging in Go.

## Structure

Zap takes an opinionated stance on logging and doesn't provide any
`printf`-style helpers. Rather than `logger.Printf("Failed to fetch URL %s
(attempt %v), sleeping %s before retry.", url, tryNum, sleepFor)`, zap
encourages the more structured

```
logger.Info("Failed to fetch URL.",
  zap.String("url", url),
  zap.Int("attempt", tryNum),
  zap.Duration("backoff", sleepFor),
)
```

This a bit more verbose, but it enables powerful ad-hoc analysis, flexible
dashboarding, and accurate message bucketing. In short, it helps you get the
most out of tools like ELK, Splunk, and Sentry. All log messages are
JSON-serialized, though PRs to support other formats are welcome.

## Performance

For applications that log in the hot path, reflection-based serialization and
string formatting are prohibitively expensive &mdash; they're CPU-intensive and
make many small allocations. Put differently, using `encoding/json` and
`fmt.Println` to log tons of `interface{}`s makes your application slow.

Zap takes a different approach. It includes a reflection-free, zero-allocation
JSON encoder, and it offers a variety of type-safe ways to add structured
context to your log messages. It strives to avoid serialization overhead and
allocations wherever possible, so collecting rich debug logs doesn't impact
normal operations.

As measured by its own [benchmarking suite][], not only is zap more
performant than comparable structured logging libraries &mdash; it's also faster
than the standard library. Like all benchmarks, take these with a grain of
salt.<sup id="anchor-versions">[1](#footnote-versions)</sup>

Log a message and 10 fields:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 1171 ns/op | 705 B/op | 2 allocs/op |
| logrus | 8410 ns/op | 3560 B/op | 67 allocs/op |
| go-kit | 7380 ns/op | 3204 B/op | 70 allocs/op |
| log15 | 20610 ns/op | 4207 B/op | 90 allocs/op |

Log a message using a logger that already has 10 fields of context:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 231 ns/op | 0 B/op | 0 allocs/op |
| logrus | 8035 ns/op | 3438 B/op | 61 allocs/op |
| go-kit | 6790 ns/op | 2486 B/op | 48 allocs/op |
| log15 | 20709 ns/op | 3543 B/op | 69 allocs/op |

Log a static string, without any context or `printf`-style formatting:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | :---: |
| :zap: zap | 223 ns/op | 0 B/op | 0 allocs/op |
| standard library | 562 ns/op | 32 B/op | 2 allocs/op |
| logrus | 2765 ns/op | 1336 B/op | 26 allocs/op |
| go-kit | 1092 ns/op | 624 B/op | 13 allocs/op |
| log15 | 5513 ns/op | 1351 B/op | 23 allocs/op |

## Development Status: Beta
Ready for adventurous users, but breaking API changes are likely.

<hr>
Released under the [MIT License](LICENSE.txt).

<sup id="footnote-versions">1</sup> In particular, note that we may be
benchmarking against slightly older versions of other libraries. Versions are
pinned in zap's [glide.lock][] file. [↩](#anchor-versions)

[doc-img]: https://godoc.org/github.com/uber-common/zap?status.svg
[doc]: https://godoc.org/github.com/uber-common/zap
[ci-img]: https://travis-ci.org/uber-common/zap.svg?branch=master
[ci]: https://travis-ci.org/uber-common/zap
[cov-img]: https://coveralls.io/repos/github/uber-common/zap/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber-common/zap?branch=master
[benchmarking suite]: https://github.com/uber-common/zap/tree/master/benchmarks
[glide.lock]: https://github.com/uber-common/zap/blob/master/glide.lock
