package buffer

import (
	"bufio"
	"encoding/binary"
	"io"
	"unsafe"
)

// DefaultBufferSize represents the default buffer size whenever the buffer size
// is not set or a negative value is presented.
const DefaultBufferSize = 1 << 24 // 16777216 bytes

// BufferedReader extended io.Reader with some convenience methods.
type BufferedReader interface {
	io.Reader
	ReadString(delim byte) (string, error)
	ReadByte() (byte, error)
}

// Reader provides a convenient way to read pgwire protocol messages
type Reader struct {
	Buffer         BufferedReader
	Msg            []byte
	MaxMessageSize int
	header         [4]byte
}

// NewReader constructs a new Postgres wire buffer for the given io.Reader
func NewReader(reader io.Reader, bufferSize int) *Reader {
	if reader == nil {
		return nil
	}

	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}

	return &Reader{
		Buffer:         bufio.NewReaderSize(reader, bufferSize),
		MaxMessageSize: bufferSize,
	}
}

// reset sets reader.Msg to exactly size, attempting to use spare capacity
// at the end of the existing slice when possible and allocating a new
// slice when necessary.
func (reader *Reader) reset(size int) {
	if reader.Msg != nil {
		reader.Msg = reader.Msg[len(reader.Msg):]
	}

	if cap(reader.Msg) >= size {
		reader.Msg = reader.Msg[:size]
		return
	}

	allocSize := size
	if allocSize < 4096 {
		allocSize = 4096
	}
	reader.Msg = make([]byte, size, allocSize)
}

func (reader *Reader) ReadMessage() (int, error) {
	nread, err := io.ReadFull(reader.Buffer, reader.header[:])
	if err != nil {
		return nread, err
	}

	size := int(binary.BigEndian.Uint32(reader.header[:]))

	if size > reader.MaxMessageSize || size < 0 {
		return nread, NewMessageSizeExceeded(reader.MaxMessageSize, size)
	}

	reader.reset(size)
	n, err := io.ReadFull(reader.Buffer, reader.Msg)
	return n, err
}

// GetUint16 returns the buffer's contents as a uint16.
func (reader *Reader) GetUint16() (uint16, error) {
	if len(reader.Msg) < 2 {
		return 0, NewInsufficientData(len(reader.Msg))
	}

	v := binary.BigEndian.Uint16(reader.Msg[:2])
	reader.Msg = reader.Msg[2:]
	return v, nil
}

// GetUint32 returns the buffer's contents as a uint32.
func (reader *Reader) GetUint32() (uint32, error) {
	if len(reader.Msg) < 4 {
		return 0, NewInsufficientData(len(reader.Msg))
	}

	v := binary.BigEndian.Uint32(reader.Msg[:4])
	reader.Msg = reader.Msg[4:]
	return v, nil
}

// GetString reads a null-terminated string.
func (reader *Reader) GetString() (string, error) {
	length, err := reader.GetUint16()
	if err != nil {
		return "", err
	}

	if len(reader.Msg) < int(length) {
		return "", NewInsufficientData(len(reader.Msg))
	}

	// Note: this is a conversion from a byte slice to a string which avoids
	// allocation and copying. It is safe because we never reuse the bytes in our
	// read buffer. It is effectively the same as: "s := string(b.Msg[:pos])"
	s := reader.Msg[:length]
	reader.Msg = reader.Msg[length+1:]
	return *((*string)(unsafe.Pointer(&s))), nil
}
