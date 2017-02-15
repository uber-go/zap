# Changelog

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
