package shadowsocks

import (
	"testing"
	"fmt"
	"os"
)

func TestParseJson (t * testing.T) {
	l := len(os.Args)
	jsonFileName := os.Args[l - 1]
	fmt.Print(jsonFileName)
	config, err := ParseConfig(jsonFileName)
	if nil != err {
		fmt.Print("Parsing json file error!\n")
		fmt.Print(err.Error())
	}
	config.Print()
}

func TestUpdateConfig (t * testing.T) {
	l := len(os.Args)
	jsonFileName := os.Args[l - 1]
	fmt.Print(jsonFileName)
	fileConfig, err := ParseConfig(jsonFileName)
	if nil != err {
		fmt.Print("Parsing json file error!\n")
		fmt.Print(err.Error())
	}
	fileConfig.Print()
	config := &Config{}
	config.Print()
	config.Update(fileConfig)
	config.Print()
	fileConfig.Print()
	fmt.Print(config.Server == fileConfig.Server)
	fmt.Print(config.ServerPort == fileConfig.ServerPort)
}

func TestParseArgs (t * testing.T) {
	ParseArgs(false).Print()
}
