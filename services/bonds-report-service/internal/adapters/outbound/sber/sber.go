package sber

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gladinov/e"

	"gopkg.in/yaml.v3"
)

type ConfigSber struct {
	Bonds string `yaml:"Bonds"`
}

type Client struct {
	Portfolio map[string]float64
}

func (c *Client) GetPortfolio() map[string]float64 {
	return c.Portfolio
}

func loadConfigSber(filename string) (_ ConfigSber, err error) {
	defer func() { err = e.WrapIfErr("load sber config error", err) }()
	var c ConfigSber
	input, err := os.ReadFile(filename)
	if err != nil {
		return ConfigSber{}, err
	}
	err = yaml.Unmarshal(input, &c)
	if err != nil {
		return ConfigSber{}, err
	}
	return c, nil
}

func processConfigSber(config ConfigSber) (map[string]float64, error) {
	retBonds := make(map[string]float64)
	bonds := strings.Split(config.Bonds, ",")
	for _, v := range bonds {
		bond := strings.Split(v, ":")
		ticker := bond[0]
		quantity, err := strconv.Atoi(bond[1])
		if err != nil {
			return nil, e.WrapIfErr("can't process sber config", err)
		}
		retBonds[ticker] = float64(quantity)
	}
	return retBonds, nil
}

func NewClient(rootPath, sberConfigPath string) (*Client, error) {
	filename := filepath.Join(rootPath, sberConfigPath)
	config, err := loadConfigSber(filename)
	if err != nil {
		return nil, fmt.Errorf("load sber config failed: %w", err)
	}
	portfolio, err := processConfigSber(config)
	if err != nil {
		return nil, fmt.Errorf("process sber config failed: %w", err)
	}
	var client Client
	client.Portfolio = portfolio
	return &client, nil
}
