package config_test

import (
	"fmt"
	"os"

	"github.com/JeremyLoy/config"
)

func Example() {

	os.Setenv("DATABASEURL", "db://")
	os.Setenv("PORT", "1234")
	os.Setenv("FEATUREFLAG", "true") // also accepts t, f, 0, 1 etc. see strconv package.
	// Double underscore for sub structs. Space separation for slices.
	os.Setenv("SUBCONFIG__IPWHITELIST", "0.0.0.0 1.1.1.1 2.2.2.2")

	var c MyConfig
	config.FromEnv().To(&c)

	// db://
	fmt.Println(c.DatabaseURL)
	// 1234
	fmt.Println(c.Port)
	// true
	fmt.Println(c.FeatureFlag)
	// [0.0.0.0 1.1.1.1 2.2.2.2] 3
	fmt.Println(c.SubConfig.IPWhitelist, len(c.SubConfig.IPWhitelist))

	// Output:
	// db://
	// 1234
	// true
	// [0.0.0.0 1.1.1.1 2.2.2.2] 3
}
