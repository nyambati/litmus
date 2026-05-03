package workspace

import (
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
)

type (
	Workspace struct {
		root         *types.AlertmanagerConfig `yaml:"root,omitempty"`
		tests        []*types.TestCase         `yaml:"tests,omitempty"`
		fragments    []*fragment.Fragment
		rootFragment *fragment.Fragment `yaml:"-"`
		logger       logrus.FieldLogger `yaml:"-"`
		dir          string             `yaml:"-"`
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
