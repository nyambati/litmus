package templates

import (
	_ "embed"
)

var (
	//go:embed files/litmus.yaml
	LitmusYAML string
	//go:embed files/README.md
	ReadmeMD string
	//go:embed files/gitattributes
	GitAtrributes string
)
