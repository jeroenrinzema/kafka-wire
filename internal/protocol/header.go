package protocol

import "github.com/jeroenrinzema/kafka-wire/internal/buffer"

type RequestHeader struct {
	Key         uint16
	Version     uint16
	Correlation uint32
	ClientID    string
}

func (header *RequestHeader) Decode(reader *buffer.Reader) (err error) {
	header.Key, err = reader.GetUint16()
	if err != nil {
		return err
	}

	header.Version, err = reader.GetUint16()
	if err != nil {
		return err
	}

	header.Correlation, err = reader.GetUint32()
	if err != nil {
		return err
	}

	header.ClientID, err = reader.GetString()
	if err != nil {
		return err
	}

	return nil
}
