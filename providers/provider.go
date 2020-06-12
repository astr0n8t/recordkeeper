package providers

import (
	"github.com/nadehi18/recordkeeper/providers/cloudflare"
)

type Provider interface {
	GetIP(string) string
	SetIP(string) bool
}

func GetProvider(name string, username string, authToken string) Provider {

	if name == "cloudflare" {
		return cloudflare.Cloudflare{username, authToken}
	}

	return cloudflare.Cloudflare{}
}
