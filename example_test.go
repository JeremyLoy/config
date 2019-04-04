package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/JeremyLoy/config"
)

type MySubConfig struct {
	IPWhitelist []string
}

type MyConfig struct {
	DatabaseURL string
	Port        int
	FeatureFlag bool
	SubConfig   MySubConfig
}

func Example() {

	os.Setenv("DATABASEURL", "db://")
	os.Setenv("PORT", "1234")
	os.Setenv("FEATUREFLAG", "true") // also accepts t, f, 0, 1 etc. see strconv package.
	// Double underscore for sub structs. Space separation for slices.
	os.Setenv("SUBCONFIG__IPWHITELIST", "0.0.0.0 1.1.1.1 2.2.2.2")

	var c MyConfig
	config.FromEnv().To(&c)

	fmt.Println(c.DatabaseURL)
	fmt.Println(c.Port)
	fmt.Println(c.FeatureFlag)
	fmt.Println(c.SubConfig.IPWhitelist, len(c.SubConfig.IPWhitelist))

	// Output:
	// db://
	// 1234
	// true
	// [0.0.0.0 1.1.1.1 2.2.2.2] 3
}

func Example_fromFileWithOverride() {
	tempFile, _ := ioutil.TempFile("", "temp")
	tempFile.Write([]byte(strings.Join([]string{"PORT=1234", "FEATUREFLAG=true"}, "\n")))
	tempFile.Close()

	os.Setenv("DATABASEURL", "db://")
	os.Setenv("PORT", "5678")

	var c MyConfig
	config.From(tempFile.Name()).FromEnv().To(&c)

	// db:// was only set in ENV
	fmt.Println(c.DatabaseURL)
	// 1234 was overridden by 5678
	fmt.Println(c.Port)
	// FeatureFlag was was only set in file
	fmt.Println(c.FeatureFlag)

	// Output:
	// db://
	// 5678
	// true
}
