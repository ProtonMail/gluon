package unleash

import (
	"sync"
	"testing"
)

var (
	instance FeatureFlagValueProvider = &NullFeatureFlagProvider{}
	syncOnce sync.Once
)

func Init(provider FeatureFlagValueProvider) {
	if testing.Testing() {
		instance = provider
		return
	}

	syncOnce.Do(func() {
		instance = provider
	})
}

func Get() FeatureFlagValueProvider {
	return instance
}
