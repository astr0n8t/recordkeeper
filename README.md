# RecordKeeper

RecordKeeper is a dynamic DNS client written in Go with support for Cloudflare DNS services.  It was built for Linux, but since it was written in Go, there is a good chance that it will work on Windows and macOS as well if needed.

## The Basics

Simply create a configuration file using any file format that [Viper](https://github.com/spf13/viper) supports (JSON, TOML, YAML, HCL, INI, envfile and Java Properties files) and place it in the current directory, "~/.config/recordkeeper" or "/etc/recordkeeper" and start RecordKeeper.

RecordKeeper by default runs continually, but can also be set to run once so that it can be used with cron or a similar task scheduler.  It supports multiple DNS entries as long as they are all on the same account, but they do not need to be on the same zone.  RecordKeeper can track the current public IP address or set a static IP if for some reason you needed to ensure a record always points to a specific address.

Right now Cloudflare is the only supported provider, but because of how RecordKeeper is structured, additional providers should be relatively easy to implement and add support for.

### Configuration

To configure RecordKeeper, you need a config file, although the following CLI arguments are supported as well:

```(NONE)

      --authToken string   The authentication token to connect to the DNS provider
      --interval int       The time in seconds to check the DNS record, set to 0 to only run once (default 60)
      --provider string    Selects a DNS provider to use (default "cloudflare")
      --username string    The username to use to connect to the DNS service

```

A typical configuration file looks like this (YAML):

```(YAML)

provider: cloudflare                      # The provider to use
username: user@example.com                # Email or username for the provider account
authToken: xxxxxxxxxxxxxxxxxxxxxxxxxxxxx  # Authentication token for the account to use the service's API
interval: 60                              # The interval in minutes to recheck for changes (default is 60) when set to 0 it only checks once then exits for use with cron
records:
  -
    name: subdomain1.example.com          # Domain of the record
    address: public                       # IP or public which retrieves the current public IP address from https://ipify.org
  -
    name: subdomain2.example.com
    address: 127.0.0.1
    ID: xxxxxxxxxxxxxxxxxxxxxxxxxxxxx     # Optional: the specific ID of the record (useful when there are multiple records with the same URL)
    zoneID: xxxxxxxxxxxxxxxxxxxxxxxxxxx   # Optional: the specific ID of the zone
```

Cloudflare User Service Keys are also supported for extra security instead of using your entire account API key.  If using one simply set the username option to "CLOUDFLARESERVICEKEY" and set the authToken to your User Service Key.
