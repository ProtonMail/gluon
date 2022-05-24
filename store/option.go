package store

type Option interface {
	config(*onDiskStore)
}

func WithCompressor(cmp Compressor) Option {
	return &withCmp{
		cmp: cmp,
	}
}

type withCmp struct {
	cmp Compressor
}

func (opt withCmp) config(store *onDiskStore) {
	store.cmp = opt.cmp
}

func WithSemaphore(sem *Semaphore) Option {
	return &withSem{
		sem: sem,
	}
}

type withSem struct {
	sem *Semaphore
}

func (opt withSem) config(store *onDiskStore) {
	store.sem = opt.sem
}
