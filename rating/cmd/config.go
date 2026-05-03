package main

type config struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	Database         databaseConfig         `yaml:"database"`
	Ingester         ingesterConfig         `yaml:"ingester"`
}

type apiConfig struct {
	Port          int    `yaml:"port"`
	Host          string `yaml:"host"`
	AdvertiseHost string `yaml:"advertiseHost"`
}

type serviceDiscoveryConfig struct {
	Consul consulConfig `yaml:"consul"`
}

type consulConfig struct {
	Address string `yaml:"address"`
}

type databaseConfig struct {
	DSN string `yaml:"dsn"`
}

type ingesterConfig struct {
	Kafka kafkaConfig `yaml:"kafka"`
}

type kafkaConfig struct {
	Address string `yaml:"address"`
}
