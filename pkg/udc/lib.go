package udc

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
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
	EnvLdap            = "DAT_LDAP"
	Latest             = "2.2"
	Version            = "1.1.1"
	ReplaceIp          = "REPLACE_ME_WITH_REAL_IP_OR_HOSTNAME"
)

var (
	UnixHome         = getHome()
	UnixConfigHome   = path.Join(UnixHome, "datera")
	ConfigSearchPath = []string{getWD(), UnixHome, UnixConfigHome, UnixSiteConfigHome}
	Configs          = []string{".datera-config", "datera-config", ".datera-config.json", "datera-config.json"}
	ExampleConfig    = fmt.Sprintf(`
{"mgmt_ip": "%s",
 "username": "admin",
 "password": "password",
 "tenant": "/root",
 "api_version": "2.2",
 "ldap": ""}
`, ReplaceIp)
	ExampleRC = fmt.Sprintf(`
# DATERA ENVIRONMENT VARIABLES
export %s=%s
export %s=admin
export %s=password
export %s=/root
export %s=2.2
export %s=
`, EnvMgmt, ReplaceIp, EnvUser, EnvPass, EnvTenant, EnvApi, EnvLdap)

	EnvHelp = map[string]string{
		EnvMgmt:   "Datera management IP address or hostname",
		EnvUser:   "Datera account username",
		EnvPass:   "Datera account password",
		EnvTenant: "Datera tenant ID. eg: SE-OpenStack",
		EnvApi:    "Datera API version. eg: 2.2",
		EnvLdap:   "Datera LDAP authentication server",
	}

	IPRE_STR = `(\d{1,3}\.){3}\d{1,3}`
	IPRE     = regexp.MustCompile(IPRE_STR)

	SIP     = regexp.MustCompile(fmt.Sprintf(`san_ip\s+?=\s+?(?P<san_ip>%s)`, IPRE_STR))
	SLG     = regexp.MustCompile(`san_login\s+?=\s+?(?P<san_login>.*)`)
	SPW     = regexp.MustCompile(`san_password\s+?=\s+?(?P<san_password>.*)`)
	TNT     = regexp.MustCompile(`datera_tenant_id\s+?=\s+?(?P<tenant_id>.*)`)
	LDP     = regexp.MustCompile(`datera_ldap_server\s+?=\s+?(?P<ldap>.*)`)
	_config *UDC

	// Flags
	Fuser   = flag.String("username", "", "Datera Account Username")
	Fpass   = flag.String("password", "", "Datera Account Password")
	Fip     = flag.String("hostname", "", "Datera Hostname/Management IP")
	Ftenant = flag.String("tenant", "", "Datera Tenant")
	Fapi    = flag.String("api-version", "", "Datera Api Version")
	Fldap   = flag.String("ldap", "", "Datera LDAP authentication server")
)

type UDC struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	MgmtIp     string `json:"mgmt_ip"`
	Tenant     string `json:"tenant"`
	ApiVersion string `json:"api_version"`
	Ldap       string `json:"ldap"`
	udcFile    string
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "    ")
	return string(s)
}

func getHome() string {
	dir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	return dir
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
			uc := path.Join(loc, f)
			if _, err := os.Stat(uc); !os.IsNotExist(err) {
				fmt.Printf("Found a Universal Datera Config File: %s\n", uc)
				return uc, nil
			}
		}
	}
	if _, err := os.Stat(CinderEtc); !os.IsNotExist(err) {
		return CinderEtc, nil
	}
	return "", fmt.Errorf("NOTFOUND: No Universal Datera Config file found")
}

func unpackConfig(c []byte, file string) (*UDC, error) {
	conf := UDC{}
	err := json.Unmarshal(c, &conf)
	if err != nil {
		return nil, err
	}
	conf.udcFile = file
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
	mLdap := LDP.FindStringSubmatch(data)
	tenant := "/root"
	ldap := ""
	if len(mTenant) > 1 {
		tenant = mTenant[1]
	}
	if len(mLdap) > 1 {
		ldap = mLdap[1]
	}
	return &UDC{MgmtIp: sanIp,
		Username:   sanLogin,
		Password:   sanPassword,
		Tenant:     tenant,
		ApiVersion: Latest,
		Ldap:       ldap}, nil

}

func getBaseConfig() (*UDC, error) {
	dat := []byte{}
	cf, err := findConfigFile()
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
	return unpackConfig(dat, cf)
}

func envOverrideConfig(cf *UDC) {
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
	if ldap, ok := os.LookupEnv(EnvLdap); ok {
		cf.Ldap = ldap
	}
}

func optOverrideConfig(cf *UDC) {
	if *Fip != "" {
		cf.MgmtIp = *Fip
	}
	if *Fuser != "" {
		cf.Username = *Fuser
	}
	if *Fpass != "" {
		cf.Password = *Fpass
	}
	if *Ftenant != "" {
		cf.Tenant = *Ftenant
	}
	if *Fapi != "" {
		cf.ApiVersion = *Fapi
	}
	if *Fldap != "" {
		cf.Ldap = *Fldap
	}
}

func defaults(cf *UDC) {
	if cf.Tenant == "" {
		cf.Tenant = "/root"
	}
	if cf.ApiVersion == "" {
		cf.ApiVersion = Latest
	}
}

func checkConfig(cf *UDC) error {
	missing := []string{}
	if cf.MgmtIp == "" {
		missing = append(missing, "mgmt_ip")
	}
	if cf.Username == "" {
		missing = append(missing, "username")
	}
	if cf.Password == "" {
		missing = append(missing, "password")
	}
	if cf.Tenant == "" {
		missing = append(missing, "tenant")
	}
	if cf.ApiVersion == "" {
		missing = append(missing, "api_version")
	}
	if len(missing) > 0 {
		return fmt.Errorf("Missing UDC keys: %s", missing)
	}
	if cf.MgmtIp == ReplaceIp {
		udcFile := cf.udcFile
		if udcFile == "" {
			udcFile = "none found"
		}
		return fmt.Errorf("You must edit your UDC config file [%s] and provide "+
			"at least the mgmt_ip of the Datera box you want to connect to", udcFile)
	}
	return nil
}

func GetConfig() (*UDC, error) {
	if _config == nil {
		flag.Parse()
		cf, err := getBaseConfig()
		if err != nil && strings.HasPrefix(err.Error(), "NOTFOUND") {
			// We don't want to error out on no UDC file found in case
			// the entire UDC config is being provided by environment
			// variables or CLI opts
			cf = &UDC{}
		} else if err != nil {
			return nil, err
		}
		envOverrideConfig(cf)
		optOverrideConfig(cf)
		defaults(cf)
		if err = checkConfig(cf); err != nil {
			return nil, err
		}
		_config = cf
	}
	return _config, nil
}

func PrintConfig() {
	// Shallow copy since config doesn't have any arrays
	cf := *_config
	cf.Password = "******"
	log.Info(prettyPrint(cf))
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

func GenConfig(style string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	var (
		cf   string
		name string
	)
	if style == "shell" {
		cf = ExampleRC
		name = "daterarc"
	} else {
		cf = ExampleConfig
		name = "datera-config.json"
	}
	fmt.Printf("Writing datera config to %s/%s\n", wd, name)
	if err := ioutil.WriteFile(name, []byte(cf), 0644); err != nil {
		return err
	}
	return nil
}
