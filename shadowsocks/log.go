package shadowsocks

import (
	"io"
	"os"
	"sync"
	"runtime"
	"strconv"
	"time"
	"path"
)

const (
	CRITICAL int = 50
	ERROR    int = 40
	WARNING  int = 30
	INFO     int = 20
	DEBUG    int = 10
	NOTSET   int = 0
)

type logging struct {
	LogFileName *string
	StderrLevel int
	LogFile *os.File
	FileMutex *sync.Mutex
	StderrMutex *sync.Mutex
	LevelString map[int]string
}

// return time, file name, line number in string
func (log *logging) Loghead(skip int) (tStr, pos string) {
	t := time.Now()
	tStr = t.Format(time.UnixDate)
	_, file, line, _  := runtime.Caller(skip)
	file = path.Base(file)
	lineStr := strconv.Itoa(line)
	pos = file + " " + lineStr
	return
}

func (log *logging) MutexWrite(wr io.Writer, mes string, mut *sync.Mutex) {
	mut.Lock()
	io.WriteString(wr, mes)
	io.WriteString(wr, "\n")
	mut.Unlock()
}

func (log *logging) Write(level int, mes string, skip int) {
	str, ok := log.LevelString[level]
	if !ok {
		str = "unknown log level " + strconv.Itoa(level)
	}
	tStr, pos := log.Loghead(skip + 1)
	mes = "[" + tStr + " " + str + " " + pos + "]" + mes
	if level >= log.StderrLevel {
		log.MutexWrite(os.Stderr, mes, log.StderrMutex)
	}
	log.MutexWrite(log.LogFile, mes, log.FileMutex)
}

func (log *logging) Debug(mes string) {
	log.Write(INFO, mes, 2)
}

// (LogInit/LogFinish) and (LogChange/LogWrite) are not multithread-safa with each other
// Don't try to call LogFinish when LogWrite still running
func (log *logging) Init(logFileName string, level int) {
	log.LevelString = map[int]string{
		CRITICAL : "CRITICAL",
		ERROR    : "ERROR   ",
		WARNING  : "WARNING ",
		INFO     : "INFO    ",
		DEBUG    : "DEBUG   ",
		NOTSET   : "NOTSET  ",
	}
	io.WriteString(os.Stderr, logFileName)
	io.WriteString(os.Stderr, "\n")
	s := logFileName
	log.LogFileName = &s
	log.LogFile, _  = os.OpenFile(s, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	log.StderrLevel = level
	log.FileMutex = &sync.Mutex{}
	log.StderrMutex = &sync.Mutex{}
}
func (log *logging) Finish() {
	log.LogFile.Close()
	log.LogFile = nil
	log.LogFileName = nil
	log.FileMutex = nil
	log.StderrMutex = nil

}

// Change LogLevel or LogFile
func (log *logging) Change(logFileName string, level int) {
	log.StderrMutex.Lock()
	log.FileMutex.Lock()
	log.LogFile.Close()
	s := logFileName
	log.LogFileName = &s
	log.LogFile, _ = os.OpenFile(s, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	log.StderrLevel = level
	log.FileMutex.Unlock()
	log.StderrMutex.Unlock()
}

var Log logging
