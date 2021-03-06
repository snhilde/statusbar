![Maintenance Badge](https://img.shields.io/badge/Maintained-yes-success)
[![PkgGoDev Doc](https://pkg.go.dev/badge/github.com/snhilde/statusbar)](https://pkg.go.dev/github.com/snhilde/statusbar/v5)
[![GoReportCard](https://goreportcard.com/badge/github.com/snhilde/statusbar)](https://goreportcard.com/report/github.com/snhilde/statusbar)
[![Build Status](https://travis-ci.com/snhilde/statusbar.svg?branch=master)](https://travis-ci.com/snhilde/statusbar)


# statusbar
`statusbar` formats and displays information on the [`dwm`](https://dwm.suckless.org/) statusbar.


You can find the current release here: https://github.com/snhilde/statusbar/tree/master/v5.


## Table of Contents
1. [Overview](#overview)
1. [Installation](#installation)
1. [Usage and Documentation](#usage-and-documentation)
1. [Modules](#modules)
1. [REST API](#rest-api)
	1. [Version 1](#version-1)
		1. [Path prefix](#path-prefix)
		1. [Ping the system](#ping-the-system)
		1. [Get list of valid endpoints](#get-list-of-valid-endpoints)
		1. [Get information about all routines](#get-information-about-all-routines)
		1. [Get information about routine](#get-information-about-routine)
		1. [Restart all routines](#restart-all-routines)
		1. [Restart routine](#restart-routine)
		1. [Modify routine's settings](#modify-routines-settings)
		1. [Stop all routines](#stop-all-routines)
		1. [Stop routine](#stop-routine)
1. [Contributing](#contributing)


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
| `sbtravisci`     | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/travisci)       | Travis CI build status  |
| `sbvolume`       | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbvolume)       | Volume percentage       |
| `sbweather`      | [PkgGoDev Doc](https://pkg.go.dev/github.com/snhilde/statusbar/sbweather)      | Weather information     |


## REST API
`statusbar` comes packaged with a REST API. This API (and all future APIs) is disabled by default. To activate it, you need to call [EnableRESTAPI](https://pkg.go.dev/github.com/snhilde/statusbar#Statusbar.EnableRESTAPI) with the port you want the microservice to listen on before running the main Statusbar engine.

The REST API makes use of the wonderful [Gin](https://gin-gonic.com/) framework. For details on adding/modifying endpoints, see the documentation in the [restapi package](https://pkg.go.dev/github.com/snhilde/statusbar/restapi).

### Version 1
#### Path prefix
`/rest/v1`

#### Ping the system
![GET Badge](https://img.shields.io/badge/-GET-brightgreen) `/ping`

Sample request:
```
curl -X GET http://localhost:1234/rest/v1/ping
```

Default response:
```
Status: 200 OK
```
```
pong
```


#### Get list of valid endpoints
![GET Badge](https://img.shields.io/badge/-GET-brightgreen) `/endpoints`

Sample request:
```
curl -X GET http://localhost:1234/rest/v1/endpoints
```

Default response:
```
Status: 200 OK
```
```
{
	"endpoints": [
		{
			"method": "GET",
			"url": "/ping",
			"description": "Ping the system."
		},
		{
			"method": "GET",
			"url": "/endpoints",
			"description": "Get a list of valid endpoints."
		},
		...
	]
}
```

Internal error:
```
Status: 500 Internal Server Error
```
```
{
	"error": "error message"
}
```


#### Get information about all routines
![GET Badge](https://img.shields.io/badge/-GET-brightgreen) `/routines`

Sample request:
```
curl -X GET http://localhost:1234/rest/v1/routines
```

Default response:
```
Status: 200 OK
```
```
{
	"routines": {
		"sbbattery": {
			"name": "Battery",
			"uptime": 35212,
			"interval": 30,
			"active": true
		},
		"sbcputemp": {
			"name": "CPU Temp",
			"uptime": 35212,
			"interval": 1,
			"active": true
		},
		...
	}
}
```


#### Get information about routine
![GET Badge](https://img.shields.io/badge/-GET-brightgreen) `/routines/{routine}`

| Parameters | Location | Description |
| ---------- | -------- | ----------- |
| `routine` | path | Routine's module name |

Sample request
```
curl -X GET http://localhost:1234/rest/v1/routines/sbfan
```

Default response
```
Status: 200 OK
```
```
{
	"sbfan": {
		"name": "Fan",
		"uptime": 242,
		"interval": 1,
		"active": true
	}
}
```

Bad request
```
Status: 400 Bad Request
```
```
{
	"error": "invalid routine"
}
```


#### Restart all routines
![PUT Badge](https://img.shields.io/badge/-PUT-blue) `/routines`

Sample request
```
curl -X PUT http://localhost:1234/rest/v1/routines
```

Default response
```
Status: 204 No Content
```


#### Restart routine
![PUT Badge](https://img.shields.io/badge/-PUT-blue) `/routines/{routine}`

| Parameters | Location | Description |
| ---------- | -------- | ----------- |
| `routine` | path | Routine's module name |

Sample request
```
curl -X PUT http://localhost:1234/rest/v1/routines/sbweather
```

Default response
```
Status: 204 No Content
```

Bad request
```
Status: 400 Bad Request
```
```
{
	"error": "invalid routine"
}
```


#### Modify routine's settings
![PATCH Badge](https://img.shields.io/badge/-PATCH-blueviolet) `/routines/{routine}`

| Parameters | Location | Description |
| ---------- | -------- | ----------- |
| `routine` | path | Routine's module name |
| `interval` | body | New interval time, in seconds |

Sample request
```
curl -X PATCH --data '{"interval": 5}' http://localhost:1234/rest/v1/routines/sbcputemp
```

Default response
```
Status: 202 Accepted
```

Bad request
```
Status: 400 Bad Request
```
```
{
	"error": "error message"
}
```


#### Stop all routines
![DELETE Badge](https://img.shields.io/badge/-DELETE-red) `/routines`

Sample request
```
curl -X DELETE http://localhost:1234/rest/v1/routines
```

Default response
```
Status: 204 No Content
```

Internal error:
```
Status: 500 Internal Server Error
```
```
{
	"error": "failure"
}
```


#### Stop routine
![DELETE Badge](https://img.shields.io/badge/-DELETE-red) `/routines/{routine}`

| Parameters | Location | Description |
| ---------- | -------- | ----------- |
| `routine` | path | Routine's module name |

Sample request
```
curl -X DELETE http://localhost:1234/rest/v1/routines/sbgithubclones
```

Default response
```
Status: 204 No Content
```

Bad request
```
Status: 400 Bad Request
```
```
{
	"error": "invalid routine"
}
```

Internal error:
```
Status: 500 Internal Server Error
```
```
{
	"error": "failure"
}
```


## Contributing
If you find a bug, please submit a pull request.
If you think there could be an improvement, please open an issue or submit a pull request with the recommended change.
Contributions and new modules are always welcome.
