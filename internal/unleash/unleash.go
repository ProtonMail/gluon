package unleash

import "github.com/ProtonMail/gluon/imap"

var CapabilityKillSwitchMap = map[string]string{
	string(imap.IDLE): `InboxBridgeImapIdleCapabilityDisabled`,
}

type FeatureFlagValueProvider interface {
	GetFlagValue(key string) bool
}

type NullFeatureFlagProvider struct{}

func (n *NullFeatureFlagProvider) GetFlagValue(_ string) bool {
	return false
}
