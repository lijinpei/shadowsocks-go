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
}

/*
func TestSerialLog(t *testing.T) {
	var log *logging = &logging{}
	log.LogInit("/tmp/test.log", DEBUG)
	for level:= 0; level < CRITICAL; level += 10 {
		for info:= 0; info < 100; info += 1 {
			log.Write(level, strconv.Itoa(info), 1)
		}
	}
	log.LogFinish()
}

func TestParallelLog(t *testing.T) {
	var log *logging = &logging{}
	var wg sync.WaitGroup
	log.LogInit("/tmp/test.log", DEBUG)
	for level:= 0; level < CRITICAL; level += 10 {
		for info:= 0; info < 100; info += 1 {
			wg.Add(1)
			go func(level int, mes string, skip int) {
				defer wg.Done()
				log.Write(level, mes, skip)
			}(level, strconv.Itoa(info), 1)
		}
	}
	wg.Wait()
	log.LogFinish()
}

func TestSerialChange(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	var log *logging = &logging{}
	log.LogInit("/tmp/test.log", DEBUG)
	for level:= 0; level < CRITICAL; level += 10 {
		for info:= 0; info < 100; info += 1 {
			log.Write(level, strconv.Itoa(info), 1)
			if r := rand.Intn(2); 0 == r {
				rFile := rand.Intn(100)
				rLevel := rand.Intn(51)
				rFileName := "/tmp/test.log." + strconv.Itoa(rFile)
				rLevel = rLevel * 10
				log.LogChange(rFileName, rLevel)
				fmt.Print("log config changed " + rFileName + " " + strconv.Itoa(rLevel) + "\n")
			}
		}
	}
	log.LogFinish()
}

func TestParallelChange(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	var log *logging = &logging{}
	var wg sync.WaitGroup
	log.LogInit("/tmp/test.log", DEBUG)
	for level:= 0; level < CRITICAL; level += 10 {
		for info:= 0; info < 100; info += 1 {
			wg.Add(1)
			go func(level int, mes string, skip int) {
				defer wg.Done()
				log.Write(level, mes, skip)
				if r := rand.Intn(2); 0 == r {
					rFile := rand.Intn(100)
					rLevel := rand.Intn(51)
					rFileName := "/tmp/test.log." + strconv.Itoa(rFile)
					rLevel = rLevel * 10
					wg.Add(1)
					go func(fFileName string, rLevel int) {
						defer wg.Done()
						log.LogChange(rFileName, rLevel)
						fmt.Print("log config changed " + rFileName + " " + strconv.Itoa(rLevel) + "\n")
					}(rFileName, rLevel)
			}
			}(level, strconv.Itoa(info), 1)
		}
	}
	wg.Wait()
	log.LogFinish()
}
*/
