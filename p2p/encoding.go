package p2p

import (
	"encoding/gob"
	"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec *GOBDecoder) Decode(r io.Reader, rpc *RPC) error {
	return gob.NewDecoder(r).Decode(rpc)
}

type NOPDecoder struct{}

func (dec *NOPDecoder) Decode(r io.Reader, rpc *RPC) error {
	peekBuf := make([]byte, 1)

	if _, err := r.Read(peekBuf); err != nil {
		return err
	}

	stream := peekBuf[0] == IncomingStream

	// incase of a stream we are not decoding what is being sent
	// We are just setting stream true so we can handle that in our logic
	fmt.Println("is stream here, ", stream)
	if stream {
		rpc.Stream = true
		return nil
	}

	buf := make([]byte, 1028)
	n, err := r.Read(buf)

	if err != nil {
		return err
	}

	rpc.Payload = buf[:n]

	return nil
}
