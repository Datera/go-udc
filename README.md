# Universal Datera Config (Golang)

The Universal Datera Config (UDC) is a config that can be specified in a
number of ways:

* JSON file with any of the following names:
    [.datera-config, datera-config, .datera-config.json, dateraconfig.json]
* The JSON file has the following configuration:
```json
     {"mgmt_ip": "1.1.1.1",
      "username": "admin",
      "password": "password",
      "tenant": "/root",
      "api_version": "2.2",
      "ldap": ""}
```
* The file can be in any of the following places.  This is also the lookup
  order for config files:
    current directory --> home directory --> home/config directory --> /etc/datera
* If no datera config file is found and a cinder.conf file is present, the
  config parser will try and pull connection credentials from the
  cinder.conf
* Instead of a JSON file, environment variables can be used.
    - DAT\_MGMT
    - DAT\_USER
    - DAT\_PASS
    - DAT\_TENANT
    - DAT\_API
    - DAT\_LDAP
* Most tools built to use the Datera Universal Config will also allow
  for providing/overriding any of the config values via command line flags.
    - -hostname
    - -username
    - -password
    - -tenant
    - -api-version
    - -ldap

## Developing with Universal Datera Config

The Golang UDC is the following struct:

```go
type UDC struct {
    Username   string `json:"username"`
    Password   string `json:"password"`
    MgmtIp     string `json:"mgmt_ip"`
    Tenant     string `json:"tenant"`
    ApiVersion string `json:"api_version"`
    Ldap       string `json:"ldap"`
}
```

To use UDC in a new python tool is very simple just add the following to
your Go main package

```go
import (
    udc "github.com/Datera/go-udc/pkg/udc"
)

func main() {
    conf, err := udc.GetConfig()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("Datera MgmtIp: %s\n", conf.MgmtIp)
}
```

If you can print the supported environment variables with the following:

```go
import (
    udc "github.com/Datera/go-udc/pkg/udc"
)

func main() {
    udc.PrintEnvs()
}
```
