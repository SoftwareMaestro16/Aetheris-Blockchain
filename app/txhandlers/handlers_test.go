package txhandlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPostHandlerReturnsHandler(t *testing.T) {
	require.NotPanics(t, func() {
		_ = NewPostHandler()
	})
}
