package main

type config struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
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
