/*
Pacakge fiplog provides log wrapper
 */
package fiplog

import (
	"strings"
	"os"
	"io"
	"time"
	"regexp"
	"fmt"
	"runtime"
	"sync"
	"fipml"
	"log"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarning
	LevelError
)
var prefix = [4]string {"DEBUG","INFO","WARNING","ERROR"}

const (
	defaultBufferSize = 1024 * 64
	defaultDateFormat = `%df{2006-01-02 15:04:05}`
	defaultPattern = `%date [%level] %file - %msg`
	dateF = "%date"
	fileF = "%file"
	levelF = "%level"
	msgF = "%msg"
)

var dateFormatR = regexp.MustCompile(`%df\{(.*)\}`)

type Config struct {
	Level Level
	//levelTo Level
	Filename  string
	BufSize int
	//size todo: rolling
	Pattern string //todo: pattern
}

type Logger struct {
	//config *Config
	mutex 		sync.Mutex
	level		Level
	//levelTo		Level
	buffer		[]byte
	bufSize		int
	//format		string
	writer		io.WriteCloser
	pattern		string
	ch			chan bool
	flushMu		sync.Mutex
	/*debugLogger *log.Logger
	infoLogger *log.Logger
	warnLogger *log.Logger
	errorLogger *log.Logger*/
}


var (
	//discard = log.New(ioutil.Discard,"",0)
	//discard io.WriteCloser = fakeio(0)

	logger *Logger
)

func GetLogger() *Logger {
	if logger == nil {
		Init()
	}
	return logger
}

func Init()  {
	if logger == nil {
		//config := &Config{LevelWarning, "stdout",4096,"%df{2006-01-02 15:04} [%level] %file - %msg"} //todo: read config file

		//wd := os.Getwd()
		fml, err := fipml.Load("fiplog.fml")
		if err != nil {
			fmt.Println("fiplog.fml not found, use default config.")
			config := &Config{LevelInfo, "",defaultBufferSize,defaultPattern}
			InitWithConfig(config)
		} else {
			InitWithFml(fml)
		}

	}
}

func InitWithFml(fml *fipml.FML) {
	config := new(Config)
	config.Filename = fml.GetString("file","")
	config.Level = getLevel(fml.GetString("level",""))
	config.BufSize = fml.GetInt("bufSize", defaultBufferSize)
	config.Pattern = fml.GetString("pattern",defaultPattern)
	logger =  InitWithConfig(config)
}

func InitWithConfig(config *Config) *Logger {
	if logger == nil {
		logger = new(Logger)
		logger.level = config.Level
		if config.BufSize <= 0 {
			logger.bufSize = defaultBufferSize
		} else {
			logger.bufSize = config.BufSize
		}
		logger.buffer = make([]byte,0,logger.bufSize)
		if strings.Contains(config.Pattern,dateF) {
			logger.pattern = strings.Replace(config.Pattern,dateF,defaultDateFormat,1)
		} else {
			logger.pattern = config.Pattern
		}
		if len(config.Filename) == 0 {
			logger.writer = os.Stdout
		} else {
			logger.writer = openFile(config.Filename)
		}

		/*out := strings.ToLower(config.out)
		switch out {
		case "stdout":
			logger.writer = os.Stdout
		case "stderr":
			logger.writer = os.Stderr
		default:
			logger.writer = openFile(out)
		}

		flag := log.Ldate|log.Ltime | log.Lshortfile

		switch config.level {
		case Debug:
			logger.debugLogger = log.New(writer, "DEBUG", flag)
			fallthrough
		case Info:
			logger.infoLogger = log.New(writer, "INFO", flag)
			fallthrough
		case Warn:
			logger.warnLogger = log.New(writer, "WARN", flag)
			fallthrough
		case Error:
			logger.errorLogger = log.New(writer, "ERROR", flag)
		}*/
	}

	return logger
}

func (l *Logger) Close() {
	l.flushAndClose(l.buffer,true)
}

func getLevel(s string) Level {
	switch strings.ToUpper(s){
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARNING":
		return LevelWarning
	case "ERROR":
		return LevelError
	}
	return LevelInfo
}

func (l *Logger) format(level Level, msg string) (formatted string) {
	formatted = l.pattern
	if dateFormatR.MatchString(formatted) {
		now := time.Now()
		dt := now.Format(dateFormatR.FindStringSubmatch(formatted)[1])
		formatted = dateFormatR.ReplaceAllString(formatted,dt)
	}
	if(strings.Contains(formatted, fileF)) {
		counter, file, line, ok := runtime.Caller(3)
		if !ok {
			file = "???"
			line = 0
		}
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}

		fs := fmt.Sprintf("<%d>%v:%v",counter,short,line)
		formatted = strings.Replace(formatted, fileF,fs,1)
	}
	if(strings.Contains(formatted,levelF)) {
		formatted = strings.Replace(formatted, levelF, prefix[level],1)
	}
	formatted = strings.Replace(formatted, msgF,msg,1)

	return
}

func (l *Logger) log(level Level, msg string) {
	//log.Println("level:",level,"l.level:",l.level,",msg:",msg)
	if(level < l.level ) {
		return
	}

	formatted := l.format(level,msg)
	if l.writer == os.Stdout {
		l.writer.Write([]byte(formatted))
	} else {
		l.mutex.Lock()
		if len(l.buffer)+len(formatted) > l.bufSize {
			go l.Flush(l.buffer)
			l.buffer = make([]byte, 0, logger.bufSize)
		}
		l.buffer = append(l.buffer, formatted...)
		len := len(formatted)
		if ( len > 0 && formatted[len-1] != '\n') {
			l.buffer = append(l.buffer, '\n')
		}
		l.mutex.Unlock()
	}

	/*now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(2)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)*/
	return
}

func (l *Logger) Debug(v...interface {}) {
	msg := fmt.Sprintln(v...)
	l.log(LevelDebug, msg)
}

func (l *Logger) Debugf(format string, v...interface {}) {
	msg := fmt.Sprintf(format,v...)
	l.log(LevelDebug, msg)
}

func (l *Logger) Info(v...interface {}) {
	msg := fmt.Sprintln(v...)
	l.log(LevelInfo,msg)
}

func (l *Logger) Infof(format string, v...interface {}) {
	msg := fmt.Sprintf(format,v...)
	l.log(LevelInfo, msg)
}

func (l *Logger) Warning(v...interface {}) {
	msg := fmt.Sprintln(v...)
	l.log(LevelWarning, msg)
}

func (l *Logger) Warningf(format string, v...interface {}) {
	msg := fmt.Sprintf(format, v...)
	l.log(LevelWarning, msg)
}

func (l *Logger) Error(v...interface {}) {
	msg := fmt.Sprintln(v...)
	l.log(LevelError,msg)
}

func (l *Logger) Errorf(format string, v...interface {}) {
	msg:= fmt.Sprintf(format,v...)
	l.log(LevelError, msg)
}

func (l *Logger) Flush(buffer []byte) {
	l.flushAndClose(buffer, false)
}

func (l *Logger) flushAndClose(buffer []byte, close bool) {
	l.flushMu.Lock()
	//log.Println("flush buffer lenght:",len(buffer))
	_,err := l.writer.Write(buffer)
	if err != nil {
		fmt.Println("Flush log error:",err)
	}
	if close {
		l.writer.Close()
	}

	l.flushMu.Unlock()
}

func openFile(name string) io.WriteCloser {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		log.Println("Open log file failed, filename:",name, ",err:",err)
		return os.Stdout
	} else {
		return file
	}
}

