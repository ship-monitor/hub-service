package config

type NATS struct {
	URL string
}

type Config struct {
	NATS NATS
}

func New() Config {
	return Config{
		NATS: NATS{
			URL: "nats://localhost:4222",
		},
	}
}
