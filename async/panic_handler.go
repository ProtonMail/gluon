package async

type PanicHandler interface {
	HandlePanic(any)
}

type NoopPanicHandler struct{}

func (n NoopPanicHandler) HandlePanic(r any) {}

func HandlePanic(panicHandler PanicHandler) {
	if panicHandler == nil {
		return
	}

	if _, ok := panicHandler.(NoopPanicHandler); ok {
		return
	}

	if _, ok := panicHandler.(*NoopPanicHandler); ok {
		return
	}

	panicHandler.HandlePanic(recover())
}
