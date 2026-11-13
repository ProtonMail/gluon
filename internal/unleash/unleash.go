package unleash

var CapabilityKillSwitchMap = map[string]string{
	"IDLE": `InboxBridgeImapIdleCapabilityDisabled`, // maps to `imap.Idle`, removed dependency due to circular import
}

type FeatureFlagValueProvider interface {
	GetFlagValue(key string) bool
}

type NullFeatureFlagProvider struct{}

func (n *NullFeatureFlagProvider) GetFlagValue(_ string) bool {
	return false
}
