package store

type Option interface {
	config(*onDiskStore)
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

type withFallback struct {
	f Fallback
}

func WithFallback(f Fallback) Option {
	return &withFallback{f: f}
}

func (opt withFallback) config(store *onDiskStore) {
	store.fallback = opt.f
}
