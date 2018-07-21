package udc

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	UnixSiteConfigHome = "/etc/datera"
	CinderEtc          = "/etc/cinder/cinder.conf"
	EnvMgmt            = "DAT_MGMT"
	EnvUser            = "DAT_USER"
	EnvPass            = "DAT_PASS"
	EnvTenant          = "DAT_TENANT"
	EnvApi             = "DAT_API"
)

var (
	UnixHome         = getHome()
	UnixConfigHome   = path.Join(UnixHome, "datera")
	ConfigSearchPath = []string{getWD(), UnixHome, UnixConfigHome, UnixSiteConfigHome}
	Configs          = []string{".datera-config", "datera-config", ".datera-config.json", "datera-config.json"}
	ExampleConfig    = `
{"mgmt_ip": "1.1.1.1",
 "username": "admin",
 "password": "password",
 "tenant": "/root",
 "api_version": "2.2"}
`
	ExampleRC = fmt.Sprintf(`\
# DATERA ENVIRONMENT VARIABLES
%s=1.1.1.1
%s=admin
%s=password
%s=/root
%s=2.2
`, EnvMgmt, EnvUser, EnvPass, EnvTenant, EnvApi)

	EnvHelp = map[string]string{
		EnvMgmt:   "Datera management IP address or hostname",
		EnvUser:   "Datera account username",
		EnvPass:   "Datera account password",
		EnvTenant: "Datera tenant ID. eg: SE-OpenStack",
		EnvApi:    "Datera API version. eg: 2.2",
	}
)

type UDC struct {
	Username   string
	Password   string
	MgmtIp     string
	Tenant     string
	ApiVersion string
}

func printEnvs() {
	fmt.Println()
	fmt.Println("DATERA ENVIRONMENT VARIABLES")
	fmt.Println("============================")
	longest := 0
	keys := make([]string, len(EnvHelp))
	for key := range EnvHelp {
		if len(key) > longest {
			longest = len(key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		buff := strings.Repeat(" ", (longest - len(key)))
		fmt.Printf("%s%s -- %s", buff, key, EnvHelp[key])
	}
	fmt.Println()
}

func getHome() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func getWD() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}

func GetConfig() *UDC {
	return nil
}
