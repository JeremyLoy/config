// Package config provides typesafe, cloud native configuration binding from environment variables or files to structs.
//
// Configuration can be done in as little as two lines:
//     var c MyConfig
//     config.FromEnv().To(&c)
//
// A field's type determines what https://golang.org/pkg/strconv/ function is called.
//
// All string conversion rules are as defined in the https://golang.org/pkg/strconv/ package.
//
// time.Duration follows the same parsing rules as https://golang.org/pkg/time#ParseDuration
//
// If chaining multiple data sources, data sets are merged.
//
// Later values override previous values.
//   config.From("dev.config").FromEnv().To(&c)
//
// Unset values remain intact or as their native zero value: https://tour.golang.org/basics/12.
//
// Nested structs/subconfigs are delimited with double underscore.
//   PARENT__CHILD
//
// Env vars map to struct fields case insensitively.
// NOTE: Also true when using struct tags.
package config
