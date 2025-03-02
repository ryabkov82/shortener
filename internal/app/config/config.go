package config

import (
	"errors"
	"flag"
	"net/url"
	"strconv"
	"strings"
)

type Config struct {
	HTTPServerAddr string
	BaseURL        string
}

func Load() *Config {

	cfg := new(Config)
	cfg.HTTPServerAddr = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080/"

	/*
		flag.StringVar(&cfg.HTTPServerAddr, "a", "localhost:8080", "Server address host:port")
		flag.StringVar(&cfg.BaseURL, "b", "localhost:8080", "Base address shortened url")
	*/

	flag.Func("a", "Server address host:port", func(flagValue string) error {

		hp := strings.Split(flagValue, ":")
		if len(hp) != 2 {
			return errors.New("need address in a form host:port")
		}
		_, err := strconv.Atoi(hp[1])
		if err != nil {
			return err
		}
		cfg.HTTPServerAddr = flagValue
		return nil
	})

	flag.Func("b", "Base address shortened url", func(flagValue string) error {

		_, err := url.Parse(flagValue)

		if err != nil {
			return err
		}
		cfg.BaseURL = flagValue
		return nil
	})

	flag.Parse()

	return cfg

}
