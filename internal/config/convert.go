package config

import (
	"fmt"

	"github.com/prometheus/alertmanager/config"
	"gopkg.in/yaml.v3"
)

func ToAMConfig(ac *AlertmanagerConfig) (*config.Config, error) {
	if ac == nil {
		return nil, fmt.Errorf("AlertmanagerConfig is nil")
	}

	data, err := yaml.Marshal(ac)
	if err != nil {
		return nil, err
	}

	return config.Load(string(data))
}
