package version

import "fmt"

type Version struct {
	Major, Minor, Patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("%02v.%02v.%02v", v.Major, v.Minor, v.Patch)
}

type Info struct {
	Name       string
	Version    Version
	Vendor     string
	SupportURL string
}
