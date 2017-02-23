# Changelog

## v1.0.0-rc.2 (21 Feb 2017)
This is the second release candidate for zap's stable release. It includes two
breaking changes.

Breaking changes:

* [#316][]: Zap's global loggers are now fully concurrency-safe
  (previously, users had to ensure that `ReplaceGlobals` was called before the
  loggers were in use). However, they must now be accessed via the `L()` and
  `S()` functions. Users can update their projects with

  ```
  gofmt -r "zap.L -> zap.L()" -w .
  gofmt -r "zap.S -> zap.S()" -w .
  ```
* [#309][] and [#317][]: RC1 was mistakenly shipped with invalid
  JSON and YAML struct tags on all config structs. This release fixes the tags
  and adds static analysis to prevent similar bugs in the future.

Bugfixes:

* [#321][]: Redirecting the standard library's `log` output now
  correctly reports the logger's caller.

Enhancements:

* [#325][] and [#333][]: Zap now transparently supports non-standard, rich
  errors like those produced by `github.com/pkg/errors`.
* [#326][]: Though `New(nil)` continues to return a no-op logger, `NewNop()` is
  now preferred. Users can update their projects with `gofmt -r 'zap.New(nil) ->
  zap.NewNop()' -w .`.
* [#300][]: Incorrectly importing zap as `github.com/uber-go/zap` now returns a
  more informative error.

Thanks to @skipor and @chapsuk for their contributions to this release.

## v1.0.0-rc.1 (14 Feb 2017)
This is the first release candidate for zap's stable release. There are multiple
breaking changes and improvements from the pre-release version. Most notably:

* **Zap's import path is now "go.uber.org/zap"** &mdash; all users will
  need to update their code.
* User-facing types and functions remain in the `zap` package. Code relevant
  largely to extension authors is now in the `zapcore` package.
* The `zapcore.Core` type makes it easy for third-party packages to use zap's
  internals but provide a different user-facing API.
* `Logger` is now a concrete type instead of an interface.
* A less verbose (though slower) logging API is included by default.
* Package-global loggers `L` and `S` are included.
* A human-friendly console encoder is included.
* A declarative config struct allows common logger configurations to be managed
  as configuration instead of code.
* Sampling is more accurate, and doesn't depend on the standard library's shared
  timer heap.

## v0.1.0-beta.1 (6 Feb 2017)
This is a minor version, tagged to allow users to pin to the pre-1.0 APIs and
upgrade at their leisure. Since this is the first tagged release, there are no
backwards compatibility concerns and all functionality is new.

Early zap adopters should pin to the 0.1.x minor version until they're ready to
upgrade to the upcoming stable release.

[#316]: https://github.com/uber-go/zap/pull/316
[#309]: https://github.com/uber-go/zap/pull/309
[#317]: https://github.com/uber-go/zap/pull/317
[#321]: https://github.com/uber-go/zap/pull/321
[#325]: https://github.com/uber-go/zap/pull/325
[#333]: https://github.com/uber-go/zap/pull/333
[#326]: https://github.com/uber-go/zap/pull/326
[#300]: https://github.com/uber-go/zap/pull/300
