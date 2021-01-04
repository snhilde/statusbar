# `statusbar` Changelog


## 5.3.0

### Enhancements
	* Switched weather provider for `sbweather` to [OpenWeather](https://openweathermap.org/)


## 5.2.2

### Breaking
	* Created v4 and v5 subdirectories to follow Semantic Import Versioning rules.


## 5.2.1

### Minor
	* Update README for REST API.
	* Handle SIGTERM signal.
	* Add go.mod and go.sum module files.


## 5.2.0

### Features
	* Added REST API.


## 5.1.1

### Minor
	* Migrated documentation to GoDocs to PkgGoDev Doc


## 5.1.0

### Features
	* Added sbgithubclones module to track daily and weekly clone counts for a given repository.

### Minor
	* Engine will now truncate module outputs that are longer than 50 characters.
	* Various module enhancements.


## 5.0.0

### Breaking
	Changed repository from `statusbar4` to `statusbar`.
	Redefined `RoutineHandler`'s method `Update()` to `Update() (bool, error)`.
	Added method `Error() string` to RoutineHandler to format error messages separately from the regular output.
	Added method `Name() string` to RoutineHandler to allow modules to set their own display name for logging.

### Enhancements
	* Errors in modules are now handled more gracefully. Modules no longer need to track their own error state; the core engine handles displaying the error as well as either attempting the update procedure again after a predetermined waiting period or stopping the routine completely.
	* Implemented normal logging and error logging.


## 4.1.2

### Enhancements
	* Updated documentation.


## 4.1.1

### Minor
	* Changes made to formatting and style as per `gofmt` and `golint` recommendations.


## 4.1.0

### Breaking
	* Moved routines and engine into one common repository.


## 4.0.0

### Breaking
	* Complete rewrite in `go`.

### Minor
	* Simpler formatting and customization.


## 3.0.0

### Enhancements
	* Added support for concurrency.
	* Made routines modular.


## 2.0.0

### Major
	* Ported to Linux.


## 1.0.0

### Major
	* Initial release (OpenBSD only).
