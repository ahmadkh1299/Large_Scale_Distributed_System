package config

type RegistryConfigBase struct {
	Type       string `yaml:"type"` // <-- The struct tag: "type" key in YAML loads into Type field
	ListenPort int    `yaml:"listenPort"`
}
