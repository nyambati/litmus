package templates

import (
	"embed"
	"fmt"
)

//go:embed files/*.yaml files/*.md
var FS embed.FS

// MustRead returns the embedded template file contents, panicking if the file is missing.
func MustRead(name string) string {
	data, err := FS.ReadFile("files/" + name)
	if err != nil {
		panic(fmt.Sprintf("template %q: %v", name, err))
	}
	return string(data)
}
