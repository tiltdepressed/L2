package parser

import (
	"flag"
	"fmt"
	"time"
)

type Options struct {
	Host    string
	Port    string
	Timeout time.Duration
}

func ParseFlags() (*Options, error) {
	host := flag.String("host", "", "Хост сервера, к которому подключаемся (обязательно)")
	port := flag.String("port", "", "Порт сервера, к которому подключаемся (обязательно)")
	timeout := flag.Duration("timeout", 10*time.Second, "Таймаут подключения (например, 5s, 200ms). По умолчанию 10s")

	flag.StringVar(host, "H", "", "alias for --host")
	flag.StringVar(port, "p", "", "alias for --port")
	flag.DurationVar(timeout, "t", 10*time.Second, "alias for --timeout")

	flag.Parse()

	if *host == "" || *port == "" {
		return nil, fmt.Errorf("host/port are required")
	}

	return &Options{
		Host:    *host,
		Port:    *port,
		Timeout: *timeout,
	}, nil
}
