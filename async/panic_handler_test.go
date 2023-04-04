package async

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type recoverHandler struct{}

func (h recoverHandler) HandlePanic(r interface{}) {
	fmt.Println("recoverHandler", r)
}

func TestPanicHandler(t *testing.T) {
	require.NotPanics(t, func() {
		defer HandlePanic(recoverHandler{})
		panic("there")
	})

	require.PanicsWithValue(t, "where", func() {
		defer HandlePanic(NoopPanicHandler{})
		panic("where")
	})

	require.PanicsWithValue(t, "everywhere", func() {
		defer HandlePanic(nil)
		panic("everywhere")
	})

	require.NotPanics(t, func() {
		defer HandlePanic(recoverHandler{})
		panic(nil)
	})

	require.NotPanics(t, func() {
		defer HandlePanic(recoverHandler{})
	})

	require.NotPanics(t, func() {
		defer HandlePanic(NoopPanicHandler{})
	})

	require.NotPanics(t, func() {
		defer HandlePanic(&NoopPanicHandler{})
	})
}
