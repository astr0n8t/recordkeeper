package cloudflare

import (
	
)

type Cloudflare struct {
	username string
	authToken string
}

func (c Cloudflare) New(username string, authToken string) {
	c.username = username
	c.authToken = authToken
}

func (c Cloudflare) GetIP(domain string) string {
	return "127.0.0.1"
}

func (c Cloudflare) SetIP(domain string) bool {
	return true
}

func (c Cloudflare) authenticate() bool {
	return true
}