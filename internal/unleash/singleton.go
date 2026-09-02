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
	if testing.Testing() && instance == nil {
		instance = &NullFeatureFlagProvider{}
		return instance
	}
	return instance
}
