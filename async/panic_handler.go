package async

type PanicHandler interface {
	HandlePanic()
}

type NoopPanicHandler struct{}

func (n NoopPanicHandler) HandlePanic() {}

func HandlePanic(panicHandler PanicHandler) {
	if panicHandler != nil {
		panicHandler.HandlePanic()
	}
}
