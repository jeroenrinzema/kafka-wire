package buffer

import (
	"testing"
)

func TestNewReaderNil(t *testing.T) {
	reader := NewReader(nil, 0)
	if reader != nil {
		t.Fatalf("unexpected result, expected reader to be nil %+v", reader)
	}
}

func TestMsgReset(t *testing.T) {
	expected := 4096

	t.Run("undefined", func(t *testing.T) {
		reader := &Reader{}
		reader.reset(expected)

		if len(reader.Msg) != expected {
			t.Errorf("unexpected reader message size %d, expected %d", len(reader.Msg), expected)
		}
	})

	t.Run("greater", func(t *testing.T) {
		reader := &Reader{
			Msg: make([]byte, 0, expected*2),
		}

		reader.reset(expected)

		if len(reader.Msg) != expected {
			t.Errorf("unexpected reader message size %d, expected %d", len(reader.Msg), expected)
		}
	})

	t.Run("smaller", func(t *testing.T) {
		reader := &Reader{
			Msg: make([]byte, 0, expected/2),
		}
		reader.reset(expected)

		if len(reader.Msg) != expected {
			t.Errorf("unexpected reader message size %d, expected %d", len(reader.Msg), expected)
		}
	})
}
