package imap

// QuotaResource represents a single resource within a quota root (e.g. STORAGE or MESSAGE).
type QuotaResource struct {
	ResourceName string
	Usage        uint32
	Limit        uint32
}

// QuotaRoot represents a named quota root and its associated resources.
type QuotaRoot struct {
	RootName  string
	Resources []QuotaResource
}
