package cloudflare

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Cloudflare struct {
	username  string
	authToken string
}

type record struct {
	name       string
	id         string
	zoneID     string
	recordType string
	proxied    bool
	ttl        int
}

func New(user string, auth string) Cloudflare {
	return Cloudflare{user, auth}
}

func (c Cloudflare) GetIP(domain string) string {
	c.sendRequest("", "", domain, "GET")
	return "127.0.0.1"
}

func (c Cloudflare) SetIP(domain string) bool {
	return true
}

func findZoneName(domain string) string {
	zoneNameSlice := strings.Split(domain, ".")

	var zoneName string
	if len(zoneNameSlice) > 1 {
		zoneName = zoneNameSlice[len(zoneNameSlice)-2] + zoneNameSlice[len(zoneNameSlice)-1]
	} else {
		zoneName = domain
	}

	return zoneName
}

func (c Cloudflare) sendRequest(zoneID string, id string, domain string, method string) {
	url := "https://api.cloudflare.com/client/v4/zones"

	if id != "" && zoneID != "" {
		url += zoneID + "/dns_records/" + id
	} else if id == "" && zoneID != "" {
		url += zoneID + "/dns_records"
	}

	httpClient := http.Client{}
	request, err := http.NewRequest(method, url, nil)
	request.Header.Add("X-Auth-Email", c.username)
	request.Header.Add("X-Auth-Key", c.authToken)
	request.Header.Add("X-Content-Type", "application/json")

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resultSlice := result["result"].([]interface{})
	resultMap := resultSlice[0].(map[string]interface{})
	fmt.Println(resultMap["name"])
}
