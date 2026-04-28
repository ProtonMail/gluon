package connectionlimiter

const (
	defaultClientLimit   = 25
	unlimitedClientLimit = 0
)

type Limits struct {
	// Normalized client name with max open sessions.
	// If we want unlimited connections for a client set the limit to 0.
	PerClient map[Client]int `json:"per_client"`

	// Max open sessions for unknown clients.
	// If we want unlimited connections for unknown clients set the limit to 0.
	UnknownLimit int `json:"unknown_limit"`
}

func NewDefaultLimits() Limits {
	return Limits{
		PerClient: map[Client]int{
			ClientAppleMail:   defaultClientLimit,
			ClientOutlook:     defaultClientLimit,
			ClientThunderbird: defaultClientLimit,
		},
		UnknownLimit: defaultClientLimit,
	}
}

func NewDefaultFallbackValues() Limits {
	return Limits{
		PerClient: map[Client]int{
			ClientAppleMail:   unlimitedClientLimit,
			ClientOutlook:     unlimitedClientLimit,
			ClientThunderbird: unlimitedClientLimit,
		},
		UnknownLimit: unlimitedClientLimit,
	}
}

func NewLimits(perClient map[Client]int, unknownLimit int) Limits {
	return Limits{
		PerClient:    perClient,
		UnknownLimit: unknownLimit,
	}
}
