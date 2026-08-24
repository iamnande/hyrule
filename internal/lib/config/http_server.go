package config

type HTTPServer struct {
	Addr string `env:"ADDR" envDefault:":8000"`
}

func LoadHTTPServer() func() (HTTPServer, error) {
	return load[HTTPServer]("HTTP_SERVER")
}
