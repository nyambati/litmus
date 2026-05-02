package workspace

import (
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
)

type (
	Workspace struct {
		root      *types.AlertmanagerConfig `yaml:"root,omitempty"`
		Tests     []*types.TestCase         `yaml:"tests,omitempty"`
		Fragments []*fragment.Fragment
		Logger    logrus.FieldLogger `yaml:"-"`
		dir       string             `yaml:"-"`
	}

	Metadata struct {
		Dir       string   `yaml:"dir,omitempty"`
		BaseFile  string   `yaml:"base_file,omitempty"`
		TestFiles []string `yaml:"test_files,omitempty"`
	}

	// groupSet collects fragment routes grouped by Group.Match labels so that
	// fragments sharing a group fold into a single sub-route under root.Route.
	groupSet struct {
		entries map[string]*groupEntry
		order   []string
	}

	groupEntry struct {
		route    *amconfig.Route
		receiver string
	}
)

// func (c *types.AlertmanagerConfig) String() (string, error) {
// 	data, err := yaml.Marshal(c)
// 	return string(data), err
// }

// func (c *AlertmanagerConfig) ConvertToAMConfigStruct() (*amconfig.Config, error) {
// 	data, err := yaml.Marshal(c)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return amconfig.Load(string(data))
// }

// func (c *AlertmanagerConfig) MarshalIndent(spaces int) (*bytes.Buffer, error) {
// 	var buff bytes.Buffer
// 	encoder := yaml.NewEncoder(&buff)
// 	encoder.SetIndent(spaces)
// 	err := encoder.Encode(c)
// 	return &buff, err
// }
