# Config
[![Documentation](https://godoc.org/github.com/JeremyLoy/config?status.svg)](http://godoc.org/github.com/JeremyLoy/config)

Manage your application config as a typesafe struct in as little as two function calls.

```go
type MyConfig struct {
	DatabaseUrl string
	Port int
	FeatureFlag bool
	...
}

var c MyConfig
config.FromEnv().To(&c)
```

## How It Works

Its just simple, pure stdlib. 

* The type of a field determines what [strconv](https://golang.org/pkg/strconv/) function is called.
* All string conversion rules are as defined in the [strconv](https://golang.org/pkg/strconv/) package
* If chaining multiple data sources, data sets are merged. 
  Later values override previous values.
    * e.g. `From("dev.config").FromEnv().To(&c)`
* Unset values remain as their native [zero value](https://tour.golang.org/basics/12) 

## Why you should use this

* Its the cloud-native way to manage config. See [12 Factor Apps](https://12factor.net/config)
* Simple:
    * only 2 lines to configure.
* Composeable:
    * Merge local files and environment variables for effortless local development.
* small:
    * only stdlib 
    * < 160 LoC
    
## Design Philosophy.

Opinionated and narrow in scope. This library is only meant to do config binding. 
Feel free to use it on its own, or alongside other libraries.  

* Only structs at the entry point. This keeps the API surface small.  

* Slices are space delimited. This matches how environment variables and commandline args are handled by the `go` cmd.

* No slices of structs. The extra complexity isn't warranted for such a niche usecase.

* No maps. The only feature of maps not handled by structs for this usecase is dynamic keys.
    * TODO

* No pointer members. If you really need one, just take the address of parts of your struct.