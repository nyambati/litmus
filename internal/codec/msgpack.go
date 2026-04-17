package codec

import (
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

// DecodeMsgPack decodes a MessagePack stream into the provided interface.
func DecodeMsgPack(r io.Reader, v interface{}) error {
	return msgpack.NewDecoder(r).Decode(v)
}

// EncodeMsgPack encodes an interface into a MessagePack stream.
func EncodeMsgPack(w io.Writer, v interface{}) error {
	return msgpack.NewEncoder(w).Encode(v)
}
