package cloudflare

import (
	
)

type Cloudflare struct {
	domains []string
	authToken string
	username string
}

func New(domains []string, authToken string, username string) Cloudflare {
	c := Cloudflare {domains, authToken, username}
	return c
}

func (c Cloudflare) GetIP() string {
	return "127.0.0.1"
}

func (c Cloudflare) SetIP(domain string) bool {
	return true
}

func (c Cloudflare) Authenticate() bool {
	return true
}