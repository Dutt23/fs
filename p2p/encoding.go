package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type GOBDecoder struct{}

func (dec *GOBDecoder) Decode(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

type NOPDecoder struct{}

func (dec *NOPDecoder) Decode(r io.Reader, msg *Message) error {
	buf := make([]byte, 1028)
	n, err := r.Read(buf)

	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
