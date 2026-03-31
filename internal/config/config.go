package cfg

import "os"

type Config struct {
	Port         string
	DoctorClient string
}

func LoadCfg() *Config {
	return &Config{Port: os.Getenv("PORT"), DoctorClient: os.Getenv("DOCTOR_SCV_URL")}
}
