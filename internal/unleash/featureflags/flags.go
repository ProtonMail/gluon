package featureflags

const (
	CommandWatcherGlobalDisabled               = "InboxBridgeGenericImapOkHeartbeatDisabled"
	CommandWatcherNonThunderbirdDisabled       = "InboxBridgeGenericImapOkHeartbeatNonThunderbirdDisabled"
	ConnectionLimiterDisabled                  = "InboxBridgeGluonConnectionLimiterDisabled"
	ConnectionLimiterDefaultLimitsDisabled     = "InboxBridgeGluonConnectionLimiterDefaultLimitsDisabled"
	ConnectionCounterConnectionsLimitDisabled  = "InboxBridgeGluonRollingCounterConnectionLimitDisabled"
	MaximumMIMEStructureDepthDisabled          = "InboxBridgeGluonMaximumMimeStructureDepthLimitDisabled"
	MaximumSearchKeyDepthDisabled              = "InboxBridgeGluonMaximumSearchKeyDepthDisabled"
	ApplySentryEventsDisabled                  = "InboxDesktopGluonApplySentryEventsDisabled"
	ContentTransferEncodingDefault7BitDisabled = "InboxDesktopGluonNoContentTransferEncoding7BitDefaultDisabled"
)
