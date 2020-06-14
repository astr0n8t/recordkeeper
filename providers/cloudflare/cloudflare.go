package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Cloudflare struct {
	username  string
	authToken string
	records   map[string]record
}

type record struct {
	name       string
	id         string
	zoneID     string
	recordType string
	proxied    bool
	ttl        int
}

func New(user string, auth string) *Cloudflare {
	return &Cloudflare{user, auth, make(map[string]record)}
}

func (c *Cloudflare) GetIP(domain string) string {

	domainRecord, exists := c.records[domain]
	if !exists {
		c.records[domain] = record{domain, "", c.getZoneID(domain), "", false, 0}
		domainRecord = c.records[domain]
	}

	fmt.Println(domainRecord)

	return "127.0.0.1"
}

func (c *Cloudflare) SetIP(domain string, address string) bool {
	return true
}

func (c *Cloudflare) getZoneID(domain string) string {
	var zoneID string
	zoneName := findZoneName(domain)
	response := c.sendRequest("", "", domain, "GET")

	var zoneData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&zoneData)
	zones := zoneData["result"].([]interface{})
	for i := range zones {
		currentZone := zones[i].(map[string]interface{})
		currentZoneName := fmt.Sprintf("%v", currentZone["name"])
		if currentZoneName == zoneName {
			zoneID = fmt.Sprintf("%v", currentZone["id"])
		}
	}

	return zoneID
}

func findZoneName(domain string) string {
	zoneNameSlice := strings.Split(domain, ".")

	var zoneName string
	if len(zoneNameSlice) > 1 {
		zoneName = zoneNameSlice[len(zoneNameSlice)-2] + "." + zoneNameSlice[len(zoneNameSlice)-1]
	} else {
		zoneName = domain
	}

	return zoneName
}

func (c *Cloudflare) sendRequest(zoneID string, id string, domain string, method string) *http.Response {
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
		panic(fmt.Errorf("cannot form CloudFlare API request"))
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		panic(fmt.Errorf("cannot connect to CloudFlare API"))
	}

	return resp
}
