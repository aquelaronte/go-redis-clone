package core_test

import (
	"bytes"
	"go-redis-clone/internal/core"
	"testing"
)

func TestDatabase(t *testing.T) {
	t.Run("store a value", func(t *testing.T) {
		core.SET([]byte("fizz"), []byte("buzz"))

		value := core.GET([]byte("fizz"))

		if !bytes.Equal(value, []byte("buzz")) {
			t.Errorf("expected buz, found %s", string(value))
		}
	})

	t.Run("delete a value", func(t *testing.T) {
		core.SET([]byte("fizz"), []byte("buzz"))

		value := core.GET([]byte("fizz"))

		if !bytes.Equal(value, []byte("buzz")) {
			t.Errorf("expected buz, found %s", string(value))
		}

		core.DEL([]byte("fizz"))

		value = core.GET([]byte("fizz"))

		if value != nil {
			t.Errorf("expected nil, found %s", string(value))
		}
	})

}
