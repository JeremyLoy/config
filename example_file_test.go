package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/JeremyLoy/config"
)

func ExampleWithFile() {
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
