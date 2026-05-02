package workspace

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LoadAssembledWorkspace loads and assembles a workspace from config.
func Load(dir string, logger logrus.FieldLogger) (*Workspace, error) {
	ws := New(dir, logger)
	if _, err := ws.Assemble(); err != nil {
		return nil, fmt.Errorf("assembling workspace: %w", err)
	}
	return ws, nil
}
