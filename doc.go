// Package config provides typesafe, cloud native configuration binding from environment variables or files to structs.
//
// Configuration can be done in as little as two lines:
//
//	var c MyConfig
//	config.FromEnv().To(&c)
//
// A field's type determines what https://pkg.go.dev/strconv function is called.
//
// All string conversion rules are as defined in the https://pkg.go.dev/strconv package.
//
// time.Duration follows the same parsing rules as https://pkg.go.dev/time#ParseDuration
//
// *url.URL follows the same parsing rules as https://pkg.go.dev/net/url#URL.Parse
// NOTE: `*url.URL` fields on the struct **must** be a pointer
//
// Any type implementing [encoding.TextUnmarshaler] is parsed with its UnmarshalText method.
//
// If chaining multiple data sources, data sets are merged.
//
// Later values override previous values.
//
//	config.From("dev.config").FromEnv().To(&c)
//
// Unset values remain intact or as their native zero value: https://tour.golang.org/basics/12.
//
// Nested structs/subconfigs are delimited with double underscore.
//
//	PARENT__CHILD
//
// Env vars map to struct fields case insensitively.
// NOTE: Also true when using struct tags.
package config
