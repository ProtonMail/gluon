package unleash

type MockFeatureFlagValueProvider struct {
	flags map[string]bool
}

func NewMockFeatureFlagValueProvider(flags map[string]bool) FeatureFlagValueProvider {
	return &MockFeatureFlagValueProvider{
		flags: flags,
	}
}

func (ff *MockFeatureFlagValueProvider) GetFlagValue(key string) bool {
	return ff.flags[key]
}
