# :zap: zap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Fast, structured, leveled logging in Go.

## Structure

Zap takes an opinionated stance on logging and doesn't provide any
`printf`-style helpers. Rather than `logger.Printf("Error %v writing logs to
%v, lost %v messages.", err, f, m)`, zap encourages the more structured

```
Logger.
  WithError(err).
  With("msgCount", m).
  With("fileName", f).
  Info("Error writing logs.")
```

This a bit more verbose, but it enables powerful ad-hoc analysis, flexible
dashboarding, and accurate message bucketing. In short, it helps you get the
most out of tools like ELK, Splunk, and Sentry. All log messages are
JSON-serialized.

## Performance

For applications that log in the hot path, reflection-based serialization and
string formatting are prohibitively expensive &mdash; they're CPU-intensive and
make many small allocations. Put differently, using `encoding/json` to log tons
of `interface{}`s makes your application slow.

Zap's API offers a variety of type-safe ways to annotate a logger's context
without incurring tons of overhead. It also offers a suite of conditional
annotations, so collecting rich debugging context doesn't impact normal
operations.

As measured by its own benchmarking suite, not only is zap more performant
than comparable structured logging libraries &mdash; it's also faster than the
standard library. Like all benchmarks, take these with a grain of salt.

Add 5 fields to the logging context, one at a time:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | ---: |
| zap | 3340 ns/op | 5713 B/op | 16 allocs/op |
| logrus | 66776 ns/op | 52646 B/op | 254 allocs/op |

Add 5 fields to the logging context as a single map:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | ---: |
| zap | 1615 ns/op | 1504 B/op | 6 allocs/op |
| logrus | 36592 ns/op | 21409 B/op | 209 allocs/op |

Log static strings, without any context or `printf`-style formatting:

| Library | Time | Bytes Allocated | Objects Allocated |
| :--- | :---: | :---: | ---: |
| zap | 328 ns/op | 32 B/op | 1 allocs/op |
| standard library | 840 ns/op | 592 B/op | 2 allocs/op |

## Development Status: Alpha

Breaking changes are certain.

<hr>
Released under the [MIT License](LICENSE.txt).

[doc-img]: https://godoc.org/github.com/uber-common/zap?status.svg
[doc]: https://godoc.org/github.com/uber-common/zap
[ci-img]: https://travis-ci.org/uber-common/zap.svg?branch=master
[ci]: https://travis-ci.org/uber-common/zap
[cov-img]: https://coveralls.io/repos/github/uber-common/zap/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber-common/zap?branch=master
