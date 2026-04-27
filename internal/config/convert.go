package config

import (
	"github.com/prometheus/alertmanager/config"
	"gopkg.in/yaml.v3"
)

func ToAMConfig(ac *AlertmanagerConfig) (*config.Config, error) {
	if ac == nil {
		return nil, nil
	}

	data, err := yaml.Marshal(ac)
	if err != nil {
		return nil, err
	}

	return config.Load(string(data))
}
