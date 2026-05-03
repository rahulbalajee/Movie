package main

type config struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	Database         databaseConfig         `yaml:"database"`
}

type apiConfig struct {
	Port int    `yaml:"port"`
	// Host is the bind address ("" = all interfaces).
	Host string `yaml:"host"`
	// AdvertiseHost is the host other services use to reach us (registered with Consul).
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
