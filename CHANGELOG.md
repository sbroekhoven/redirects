# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.8] - 2023-06-15

### Changed
- Improved handling of the "Location" header to be case-insensitive
- Updated `redirectURL` retrieval to use `resp.Header.Values("Location")` for case-insensitive matching
- Refactored header setting to ensure case-insensitive behavior when creating the "Location" header


## [1.0.7] - 2023-06-14

### Added
- New `ensureHTTPPrefix` function to automatically add "http://" to URLs if missing
- Case-insensitive checking for URL prefixes

### Changed
- Improved URL validation in `validateURL` function
- Updated error handling for empty URL inputs

### Fixed
- Bug in URL parsing that caused some valid URLs to be rejected

## [1.0.6] - 2023-06-01

### Added
- Initial implementation of URL validation function

### Changed
- Refactored error handling in core functions

### Deprecated
- Old URL parsing method (to be removed in v1.1.0)

## [1.0.5] - 2023-05-15

### Security
- Updated dependencies to address potential vulnerabilities

[1.0.7]: https://github.com/username/repo/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/username/repo/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/username/repo/releases/tag/v1.0.5