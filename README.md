# Config
[![PkgGoDev](https://pkg.go.dev/badge/github.com/JeremyLoy/config)](https://pkg.go.dev/github.com/JeremyLoy/config)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)
[![Build Status](https://github.com/JeremyLoy/config/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/JeremyLoy/config/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/JeremyLoy/config)](https://goreportcard.com/report/github.com/JeremyLoy/config)
[![Coverage Status](https://coveralls.io/repos/github/JeremyLoy/config/badge.svg?branch=master)](https://coveralls.io/github/JeremyLoy/config?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/JeremyLoy/config.svg)](https://github.com/JeremyLoy/config/issues)
[![license](https://img.shields.io/github/license/JeremyLoy/config.svg?maxAge=2592000)](https://github.com/JeremyLoy/config/LICENSE)
[![Release](https://img.shields.io/github/release/JeremyLoy/config.svg?label=Release)](https://github.com/JeremyLoy/config/releases)

Manage your application config as a typesafe struct in as little as two function calls.

```go
type MyConfig struct {
	DatabaseUrl string `config:"DATABASE_URL"`
	FeatureFlag bool   `config:"FEATURE_FLAG"`
	Port        int // tags are optional. PORT is assumed
	...
}

var c MyConfig
err := config.FromEnv().To(&c)
```

## How It Works

It's just simple, pure stdlib. 

* A field's type determines what [strconv](https://pkg.go.dev/strconv) function is called.
* All string conversion rules are as defined in the [strconv](https://pkg.go.dev/strconv) package
* `time.Duration` follows the same parsing rules as [time.ParseDuration](https://pkg.go.dev/time#ParseDuration)
* `*net.URL` follows the same parsing rules as [url.Parse](https://pkg.go.dev/net/url#URL.Parse)
  * NOTE: `*net.URL` fields on the struct **must** be a pointer
* If chaining multiple data sources, data sets are merged. 
  Later values override previous values.
  ```go
  config.From("dev.config").FromEnv().To(&c)
  ```
    
* Unset values remain intact or as their native [zero value](https://tour.golang.org/basics/12) 
* Nested structs/subconfigs are delimited with double underscore 
    * e.g. `PARENT__CHILD`
* Env vars map to struct fields case insensitively
    * NOTE: Also true when using struct tags.
* Any errors encountered are aggregated into a single error value
    * the entirety of the struct is always attempted
    * failed conversions (i.e. converting "x" to an int) and file i/o are the only sources of errors
        * missing values are not errors

## Why you should use this

* It's the cloud-native way to manage config. See [12 Factor Apps](https://12factor.net/config)
* Simple:
    * only 2 lines to configure.
* Composeable:
    * Merge local files and environment variables for effortless local development.
* small:
    * only stdlib 
    * < 180 LoC
    
## Design Philosophy

Opinionated and narrow in scope. This library is only meant to do config binding. 
Feel free to use it on its own, or alongside other libraries.  

* Only structs at the entry point. This keeps the API surface small.  

* Slices are space delimited. This matches how environment variables and commandline args are handled by the `go` cmd.

* No slices of structs. The extra complexity isn't warranted for such a niche usecase.

* No maps. The only feature of maps not handled by structs for this usecase is dynamic keys.

* No pointer members. If you really need one, just take the address of parts of your struct.
  * One exception is `*url.URL`, which is explicitly a pointer for ease of use, matching the `url` package conventions
