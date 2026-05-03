package workspace

import (
	"fmt"

	"github.com/nyambati/litmus/internal/config"
	"github.com/sirupsen/logrus"
)

// LoadAssembledWorkspace loads and assembles a workspace from config.
func Load(cfg *config.LitmusConfig, logger logrus.FieldLogger) (*Workspace, error) {
	ws := New(cfg, logger)
	if err := ws.Assemble(); err != nil {
		return nil, fmt.Errorf("assembling workspace: %w", err)
	}
	if err := ws.loadRegressionState(); err != nil {
		ws.logger.Warnf("loading regression state: %v", err)
	}
	return ws, nil
}
