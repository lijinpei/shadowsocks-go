package shadowsocks

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"getopt"
)

type Config struct {
	Server     string    `json:"server"`
	ServerPort int         `json:"server_port"`
	Local      string      `json:"local_address"`
	LocalPort  int         `json:"local_port"`
	Password   string      `json:"password"`
	Method     string      `json:"method"` // encryption method
	Auth       bool        `json:"one_time_auth"`   // one time auth
	Verbose    int 	       `json:"verbose"`
	FastOpen   bool        `json:"fast_open"`
	Workers    int         `json:"workers"`
	MngAdr     string      `json:"manager_address"`
	User       string      `json:"user"`
	ForbIP     string      `json:"forbidden_ip"`
	Daemon     string      `json:"daemon"`
	PidFile    string      `json:"pid-file"`
	LogFile    string      `json:"log-file"`
	PreIPV6    bool        `json:"prefer_ipv6"`
	PortPassword map[string]string `json:"port_password"`
	Timeout      int               `json:"timeout"`
}

func (config *Config) Print() {
	fmt.Printf("%+v\n", *config)
}

func ParseConfig(path string) (config *Config, err error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		fmt.Print("Error Opening File")
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	config = &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return
}

func ParseArgs(lsLocal bool) {
	config := &Config{}
	config.InitDefault()
	if optget.IsSet('c') {
		
	}
}

