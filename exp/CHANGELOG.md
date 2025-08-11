# Changelog
All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.3.0 - 22 Oct 2024

Breaking changes:
* [#1339][]: zapslog: Drop `HandlerOptions` in favor of `HandlerOption`,
  which uses the functional options pattern.
* [#1339][]: zapslog: Rename `AddSource` option to `WithCaller`.

Enhancements:
* [#1339][]: zapslog: Record stack traces for error logs or higher.
  The new `AddStackAt` option changes this level.

Bug fixes:
* [#1344][], [#1408][]: zapslog: Comply fully with `slog.Handler` contract.
  This includes ignoring empty `Attr`s, inlining `Group`s with empty names,
  and omitting groups with no attributes.

[#1344]: https://github.com/uber-go/zap/pull/1344
[#1339]: https://github.com/uber-go/zap/pull/1339
[#1408]: https://github.com/uber-go/zap/pull/1408

Thanks to @zekth and @arukiidou for their contributions to this release.

## 0.2.0 - 9 Sep 2023

Breaking changes:
* [#1315][]: zapslog: Drop HandlerOptions.New in favor of just the NewHandler constructor.
* [#1320][], [#1338][]: zapslog: Drop support for golang.org/x/exp/slog in favor of log/slog released in Go 1.21.

[#1315]: https://github.com/uber-go/zap/pull/1315
[#1320]: https://github.com/uber-go/zap/pull/1320
[#1338]: https://github.com/uber-go/zap/pull/1338

## 0.1.0 - 1 Aug 2023

Initial release of go.uber.org/zap/exp.
This submodule contains experimental features for Zap.
Features incubated here may be promoted to the root Zap module,
but it's not guaranteed.
