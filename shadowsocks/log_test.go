package shadowsocks

import (
	"testing"
	"sync"
)

func TestSerialLog(t *testing.T) {
	var log *logging = &logging{}
	log.LogInit("/tmp/test.log", DEBUG)
	for level:= 0; level < CRITICAL; level += 10 {
		for info:= 0; info < 100; info += 1 {
			log.Write(level, "67890", 1)
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
			}(level, "123456", 1)
		}
	}
	wg.Wait()
	log.LogFinish()
}

func TestSerialChange(t *testing.T) {
}

func TestParallelChange(t *testing.T) {
}

