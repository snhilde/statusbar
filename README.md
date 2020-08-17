![Version Badge](https://img.shields.io/badge/Version-4.1-informational)
![Maintenance Badge](https://img.shields.io/badge/Maintained-yes-success)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/snhilde/statusbar4)
[![GoReportCard example](https://goreportcard.com/badge/github.com/snhilde/statusbar4)](https://goreportcard.com/report/github.com/snhilde/statusbar4)

# statusbar4
This package displays various information on the dwm statusbar.

Usage guidelines and documentation hosted at [GoDoc](https://godoc.org/github.com/snhilde/statusbar4).

This is the framework that controls the modular routines for calculating, formatting, and displaying information on the statusbar.
For modules currently integrated with this framework, see [sb4routines](https://godoc.org/github.com/snhilde/sb4routines).

## Changelog
### 4.1.1
Changes made to formatting and style as per `gofmt` and `golint` recommendations.

### 4.1
Moved routines and engine into one common repository

### 4.0
Complete rewrite in `go`, modular routines, concurrency, simpler formatting and customization

### 3.0
Added support for concurrency, made routines modular

### 2.0
Ported to Linux

### 1.0
Initial release, OpenBSD only
