package udc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"
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
	LATEST             = "2.2"
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

	IPRE_STR = `(\d{1,3}\.){3}\d{1,3}`
	IPRE     = regexp.MustCompile(IPRE_STR)

	SIP = regexp.MustCompile(fmt.Sprintf(`san_ip\s+?=\s+?(?P<san_ip>%s)`, IPRE_STR))
	SLG = regexp.MustCompile(`san_login\s+?=\s+?(?P<san_login>.*)`)
	SPW = regexp.MustCompile(`san_password\s+?=\s+?(?P<san_password>.*)`)
	TNT = regexp.MustCompile(`datera_tenant_id\s+?=\s+?(?P<tenant_id>.*)`)
)

type UDC struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	MgmtIp     string `json:"mgmt_ip"`
	Tenant     string `json:"tenant"`
	ApiVersion string `json:"api_version"`
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

func findConfigFile() (string, error) {
	for _, loc := range ConfigSearchPath {
		for _, f := range Configs {
			if _, err := os.Stat(path.Join(loc, f)); !os.IsNotExist(err) {
				return f, nil
			}
		}
	}
	if _, err := os.Stat(CinderEtc); !os.IsNotExist(err) {
		return CinderEtc, nil
	}
	return "", fmt.Errorf("No Universal Datera Config file found")
}

func unpackConfig(c []byte) (*UDC, error) {
	conf := UDC{}
	err := json.Unmarshal(c, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func readCinderConf() (*UDC, error) {
	data := ""
	foundIndex := 0
	foundLastIndex := -1
	cdat, err := ioutil.ReadFile(CinderEtc)
	lines := strings.Split(string(cdat), "\n")
	if err != nil {
		return nil, err
	}
	for index, line := range lines {
		if "[datera]" == strings.TrimSpace(line) {
			foundIndex = index
			break
		}
	}
	for index, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			foundLastIndex = index + foundIndex
			break
		}
	}
	data = strings.Join(lines[foundIndex:foundLastIndex], "")
	sanIp := SIP.FindStringSubmatch(data)[1]
	sanLogin := SLG.FindStringSubmatch(data)[1]
	sanPassword := SPW.FindStringSubmatch(data)[1]
	mTenant := TNT.FindStringSubmatch(data)
	tenant := ""
	if len(mTenant) > 1 {
		tenant = mTenant[1]
	}
	return &UDC{MgmtIp: sanIp,
		Username:   sanLogin,
		Password:   sanPassword,
		Tenant:     tenant,
		ApiVersion: LATEST}, nil

}

func getBaseConfig() (*UDC, error) {
	cf, err := findConfigFile()
	if err != nil {
		return nil, err
	}
	dat := []byte{}
	if err != nil {
		return nil, err
	}
	if cf == CinderEtc {
		return readCinderConf()
	} else {
		dat, err = ioutil.ReadFile(cf)
	}
	if err != nil {
		return nil, err
	}
	return unpackConfig(dat)
}

func envOverrideConfig(cf *UDC) {
	//EnvMgmt, EnvUser, EnvPass, EnvTenant, EnvApi
	if mgt, ok := os.LookupEnv(EnvMgmt); ok {
		cf.MgmtIp = mgt
	}
	if usr, ok := os.LookupEnv(EnvUser); ok {
		cf.Username = usr
	}
	if pass, ok := os.LookupEnv(EnvPass); ok {
		cf.Password = pass
	}
	if ten, ok := os.LookupEnv(EnvTenant); ok {
		cf.Tenant = ten
	}
	if api, ok := os.LookupEnv(EnvApi); ok {
		cf.ApiVersion = api
	}
}

func GetConfig() (*UDC, error) {
	cf, err := getBaseConfig()
	if err != nil {
		return nil, err
	}
	envOverrideConfig(cf)
	return cf, nil
}

func PrintEnvs() {
	fmt.Println()
	fmt.Println("DATERA ENVIRONMENT VARIABLES")
	fmt.Println("============================")
	longest := 0
	keys := make([]string, 0)
	for key := range EnvHelp {
		if len(key) > longest {
			longest = len(key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		buff := strings.Repeat(" ", (longest - len(key)))
		fmt.Printf("%s%s -- %s\n", buff, key, EnvHelp[key])
	}
	fmt.Println()
}
