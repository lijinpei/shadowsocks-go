package shadowsocks

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"github.com/pborman/getopt"
	"reflect"
)

type Config struct {
	Server     *string      `json:"server"`
	ServerPort *int         `json:"server_port"`
	Local      *string      `json:"local_address"`
	LocalPort  *int         `json:"local_port"`
	Password   *string      `json:"password"`
	Method     *string      `json:"method"` // encryption method
	OneTA       *bool        `json:"one_time_auth"`   // one time auth
	Verbose    *int         `json:"verbose"`
	FastOpen   *bool        `json:"fast_open"`
	Workers    *int         `json:"workers"`
	MngAdr     *string      `json:"manager_address"`
	User       *string      `json:"user"`
	ForbIP     *[]string      `json:"forbidden_ip"`
	Daemon     *string      `json:"daemon"`
	PidFile    *string      `json:"pid-file"`
	LogFile    *string      `json:"log-file"`
	PreIPV6    *bool        `json:"prefer_ipv6"`
	PortPassword *map[string]string `json:"port_password"`
	Timeout      *int               `json:"timeout"`
}

func (config *Config) Init() {
	// Default config value goes here
}

func (config *Config) Print() {
//	fmt.Printf("%+v\n", *config)
	fmt.Println()
	bytes, _ := json.MarshalIndent(*config, "", "    ")
	fmt.Print(string(bytes))
	fmt.Println()
}

func (c1 *Config) Update(c2 *Config) {
	c1Val := reflect.ValueOf(c1).Elem()
	c2Val := reflect.ValueOf(c2).Elem()

	for i := 0; i < c2Val.NumField(); i++ {
		c1Field := c1Val.Field(i)
		c2Field := c2Val.Field(i)

		if c2Field.IsNil() {
			continue
		}
		c1Field.Set(c2Field)
	//	reflect.ValueOf(&c1Field).Elem().Set(reflect.ValueOf(c2Field))

	}
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
	config.Init()
	if getopt.IsSet('c') {
		configFileName := "cnofig.json"
		getopt.StringVar(&configFileName, 'c')
		fileConfig, _ := ParseConfig(configFileName)
		config.Update(fileConfig)
	}

	getopt.IntVar(config.ServerPort, 'p')
	getopt.StringVar(config.Password, 'k')
	getopt.IntVar(config.LocalPort, 'l')
	getopt.StringVar(config.Server, 's')
	getopt.StringVar(config.Method, 'm')
	getopt.StringVar(config.Local, 'b')
	v := getopt.GetCount('v') + getopt.GetCount("Verbose")
	fmt.Print(v)
	getopt.BoolVar(config.OneTA, 'a')
	getopt.IntVar(config.Timeout, 't')
	getopt.BoolVarLong(config.FastOpen, "fast-open", 0)
	getopt.IntVarLong(config.Workers, "workers", 0)
	getopt.StringVarLong(config.MngAdr, "manager-address", 0)
	getopt.StringVarLong(config.User, "user", 0)
	getopt.ListVarLong(config.ForbIP, "forbidden", 0)
}
