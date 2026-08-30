package config

import "time"

type HTTPServer struct {
	Addr string `env:"ADDR" envDefault:":8000"`

	// ReadHeaderTimeout bounds slowloris-style connections that trickle in
	// headers. ReadTimeout/WriteTimeout bound the request/response body.
	// IdleTimeout bounds a keep-alive connection sitting idle between
	// requests - left at zero (Go's default), it falls back to
	// ReadTimeout, and a ReadTimeout of zero means no limit at all, so an
	// idle client can hold its goroutine open indefinitely.
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" envDefault:"5s"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`
}

func LoadHTTPServer() func() (HTTPServer, error) {
	return Load[HTTPServer]("HTTP_SERVER")
}
