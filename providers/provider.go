package providers

import (
	"github.com/nadehi18/recordkeeper/providers/cloudflare"
)

type provider interface {
	GetIP() string
	SetIP() bool
	Authenticate() bool
}

func GetProvider(name string) provider {
	if name == "cloudflare" {
		p := cloudflare.Cloudfare {}
	}

	return p
}