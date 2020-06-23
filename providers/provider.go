/*
Package providers defines the provider interface which defines that
every provider needs to implement UpdateEntry and SetIP functions.  The provider
will then be added as an option in the GetProvider function which returns a
new object of the specified provider.
*/
package providers

import (
	"github.com/nadehi18/recordkeeper/providers/cloudflare"
	"github.com/nadehi18/recordkeeper/record"
)

// Provider interface defines what functions a provider struct should have
type Provider interface {
	UpdateEntry(*record.Entry)
	SetIP(string, *record.Entry) bool
}

// GetProvider returns a new object of the given provider type
func GetProvider(name string, username string, authToken string) Provider {

	if name == "cloudflare" {
		return cloudflare.New(username, authToken)
	}

	return &cloudflare.Cloudflare{}
}
