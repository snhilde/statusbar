![Version Badge](https://img.shields.io/badge/Version-5.0.0-informational)
![Maintenance Badge](https://img.shields.io/badge/Maintained-yes-success)
[![PkgGoDev Doc](https://pkg.go.dev/badge/github.com/snhilde/statusbar)](https://pkg.go.dev/github.com/snhilde/statusbar)
[![GoReportCard example](https://goreportcard.com/badge/github.com/snhilde/statusbar)](https://goreportcard.com/report/github.com/snhilde/statusbar)

<span class="bg-blue text-white rounded-1 px-2 py-1" style="text-transform: uppercase">get</span>

# statusbar
`statusbar` formats and displays information on the [`dwm`](https://dwm.suckless.org/) statusbar.


## Table of Contents
1. [Overview](#overview)
1. [Installation](#installation)
1. [Usage and Documentation](#usage-and-documentation)
1. [Modules](#modules)
1. [Contributing](#contributing)
1. [Changelog](#changelog)
	1. [5.0.0](#500)
	1. [4.1.2](#412)
	1. [4.1.1](#411)
	1. [4.1](#410)
	1. [4.0](#400)
	1. [3.0](#300)
	1. [2.0](#200)
	1. [1.0](#100)


## Overview
`statusbar` is an engine for handling the various components that make up the statusbar. The components are modular routines that handle their own logic for gathering and formatting data, with each routine run in its own thread. The engine triggers the routines to run their update process according to the time interval set by the caller and gathers the individual pieces of data for printing on the statusbar.

Integrating a custom module is very simple. See [Modules](#modules) for more information.


## Installation
`statusbar` is a package, not a stand-alone program. To download the package, you can use gotools in this way:
```
go get github.com/snhilde/statusbar
```
That will also pull in the repository's modules for quick activation.


## Usage and Documentation
To get up and running with this package, follow these steps:
1. [Create a new statusbar object.](https://pkg.go.dev/github.com/snhilde/statusbar#New)
1. [Add routines to the statusbar.](https://pkg.go.dev/github.com/snhilde/statusbar#Statusbar.Append)
1. [Run the engine.](https://pkg.go.dev/github.com/snhilde/statusbar#Statusbar.Run)

You can find the complete documentation and usage guidelines at [pkg.go.dev](https://pkg.go.dev/github.com/snhilde/statusbar). The docs also include an example detailing the steps above.


## Modules
`statusbar` is modular by design, and it's simple to build and integrate modules; you only have to implement [a few methods](https://pkg.go.dev/github.com/snhilde/statusbar#RoutineHandler).

This repository includes these modules to get up and running quickly:

| Module           | Documentation                                                                  | Major usage             |
| ---------------- | ------------------------------------------------------------------------------ | ----------------------- |
| `sbbattery`      | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbbattery)      | Battery usage           |
| `sbcputemp`      | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbcputemp)      | CPU temperature         |
| `sbcpuusage`     | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbcpuusage)     | CPU usage               |
| `sbdisk`         | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbdisk)         | Filesystem usage        |
| `sbfan`          | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbfan)          | Fan speed               |
| `sbgithubclones` | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbgithubclones) | Github repo clone count |
| `sbload`         | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbload)         | System load averages    |
| `sbnetwork`      | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbnetwork)      | Network usage           |
| `sbnordvpn`      | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbnordvpn)      | NordVPN status          |
| `sbram`          | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbram)          | RAM usage               |
| `sbtime`         | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbtime)         | Current date/time       |
| `sbtodo`         | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbtodo)         | TODO list display       |
| `sbvolume`       | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbvolume)       | Volume percentage       |
| `sbweather`      | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbweather)      | Weather information     |


## Contributing
If you find a bug, please submit a pull request.
If you think their could be an improvement, please submit an issue or a pull request with the recommended change.
Contributions of new modules are also very welcome.


## Changelog
### 5.0.0
* **API Break**: Changed repository from `statusbar4` to `statusbar`.
* **API Break**: Redefined `RoutineHandler`'s method `Update()` to `Update() (bool, error)`.
* **API Break**: Added method `Error() string` to RoutineHandler to format error messages separately from the regular output.
* **API Break**: Added method `Name() string` to RoutineHandler to allow modules to set their own display name for logging.
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
