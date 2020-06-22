package providers

import (
	"github.com/nadehi18/recordkeeper/providers/cloudflare"
	"github.com/nadehi18/recordkeeper/record"
)

type Provider interface {
	GetIP(record.Entry) string
	SetIP(string, record.Entry) bool
}

func GetProvider(name string, username string, authToken string) Provider {

	if name == "cloudflare" {
		return cloudflare.New(username, authToken)
	}

	return &cloudflare.Cloudflare{}
}
