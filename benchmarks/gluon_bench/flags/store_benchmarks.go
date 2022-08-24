package flags

import "flag"

var (
	Store          = flag.String("store", "default", "Name of the storage implementation to benchmark. Defaults to regular on disk storage by default.")
	StoreWorkers   = flag.Uint("store-workers", 1, "Number of concurrent workers for store operations.")
	StoreItemCount = flag.Uint("store-item-count", 1000, "Number of items to generate in the store benchmarks.")
	StoreItemSize  = flag.Uint("store-item-size", 15*1024*1024, "Number of items to generate in the store benchmarks.")
)
