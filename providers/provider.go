package providers

import (
	"github.com/nadehi18/recordkeeper/providers/cloudflare"
)

type Provider interface {
	New (string, string)
	GetIP(string) string
	SetIP(string) bool
}

func GetProvider(name string, username string, authToken string) Provider {
	p := selectProvider(name)

	p.New(username, authToken)

	return p
}


func selectProvider(name string) Provider {
	var p Provider
	if name == "cloudflare" {
		p = cloudflare.Cloudflare {}
	}
	return p
}