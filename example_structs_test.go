package config_test

type MySubConfig struct {
	IPWhitelist []string
}

type MyConfig struct {
	DatabaseURL string
	Port        int
	FeatureFlag bool
	SubConfig   MySubConfig
}
