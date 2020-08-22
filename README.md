![Version Badge](https://img.shields.io/badge/Version-4.1.2-informational)
![Maintenance Badge](https://img.shields.io/badge/Maintained-yes-success)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/snhilde/statusbar4)
[![GoReportCard example](https://goreportcard.com/badge/github.com/snhilde/statusbar4)](https://goreportcard.com/report/github.com/snhilde/statusbar4)


# statusbar4
`statusbar4` formats and displays information for the [`dwm`](https://dwm.suckless.org/) statusbar.


## Table of Contents
1. [Overview](#overview)
1. [Installation](#installation)
1. [Usage and Documentation](#usage-and-documentation)
1. [Included modules](#included-modules)
1. [Changelog](#changelog)
	1. [4.1.2](#412)
	1. [4.1.1](#411)
	1. [4.1](#41)
	1. [4.0](#40)
	1. [3.0](#30)
	1. [2.0](#20)
	1. [1.0](#10)


## Overview
`statusbar4` is an engine for handling the various components that make up the statusbar. The components are modular routines that handle their own logic for gathering and formatting data, with each routine run in its own thread. The engine triggers the routines to run their update process according to the time interval set by the caller and gathers the individual pieces of data for printing on the statusbar.

Integrating a custom module is very simple. See [the section on modules](#included-modules) for more information.


## Installation
`statusbar4` is a package, not a stand-alone program. To download the package, you can use gotools like this:
```
go install github.com/snhilde/statusbar4
```
That will also pull in the repository's modules for quick activation.


## Usage and Documentation
To get up and running with this package, follow these steps:
1. [Create a new statusbar object.](https://godoc.org/github.com/snhilde/statusbar4#New)
1. [Add routines to the statusbar.](https://godoc.org/github.com/snhilde/statusbar4#Statusbar.Append)
1. [Run the engine.](https://godoc.org/github.com/snhilde/statusbar4#Statusbar.Run)

You can find the complete documentation and usage guidelines at [GoDoc](https://godoc.org/github.com/snhilde/statusbar4). The docs also include an example detailing the steps above.


## Included modules
`statusbar4` is modular by design, and it's simple to build and integrate modules; you only have to implement [two methods](https://godoc.org/github.com/snhilde/statusbar4#RoutineHandler).

This repository includes these modules to get up and running quickly:

| Module       | Documentation                                                            | Major usage          |
| ------------ | ------------------------------------------------------------------------ | -------------------- |
| `sbbattery`  | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbbattery)  | Battery usage        |
| `sbcputemp`  | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbcputemp)  | CPU temperature      |
| `sbcpuusage` | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbcpuusage) | CPU usage            |
| `sbdisk`     | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbdisk)     | Filesystem usage     |
| `sbfan`      | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbfan)      | Fan speed            |
| `sbload`     | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbload)     | System load averages |
| `sbnetwork`  | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbnetwork)  | Network usage        |
| `sbnordvpn`  | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbnordvpn)  | NordVPN status       |
| `sbram`      | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbram)      | RAM usage            |
| `sbtime`     | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbtime)     | Current date/time    |
| `sbtodo`     | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbtodo)     | TODO list display    |
| `sbvolume`   | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbvolume)   | Volume percentage    |
| `sbweather`  | [GoDoc docs](https://godoc.org/github.com/snhilde/statusbar4/sbweather)  | Weather information  |


## Changelog
### 5.0.0
* **API Break** Changed repository from `statusbar4` to `statusbar`.
* **API Break** Redefined `RoutineHandler`'s method `Update()` to `Update() (bool, error)`.
* Errors in modules are now handled more gracefully. Modules no longer need to track their own error state; the core engine handles displaying the error as well as either attempting the update procedure again after a predetermined waiting period or stopping the routine completely.
* Implemented normal logging and error logging.

### 4.1.2
* Updated documentation.

### 4.1.1
* Changes made to formatting and style as per `gofmt` and `golint` recommendations.

### 4.1.0
* Moved routines and engine into one common repository.

### 4.0.0
* Complete rewrite in `go`.
* Simpler formatting and customization.

### 3.0.0
* Added support for concurrency.
* Made routines modular.

### 2.0.0
* Ported to Linux.

### 1.0.0
* Initial release (OpenBSD only).
