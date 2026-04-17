package codec

import (
	"gopkg.in/yaml.v3"
	"io"
)

// DecodeYAML decodes a YAML stream into the provided interface.
func DecodeYAML(r io.Reader, v interface{}) error {
	return yaml.NewDecoder(r).Decode(v)
}

// EncodeYAML encodes an interface into a YAML stream.
func EncodeYAML(w io.Writer, v interface{}) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(v)
}
