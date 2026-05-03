package fixtures

import (
	"embed"
	"fmt"
)

//go:embed fragment/*.yaml tests/*.yaml workspace/*.yaml workspace/tests/*.yaml workspace/tests/sub/*.yaml
var FS embed.FS

func MustRead(name string) string {
	data, err := FS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("fixture %q: %v", name, err))
	}
	return string(data)
}
