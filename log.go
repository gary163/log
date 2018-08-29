package log

import (
	"time"
	"sync"
	"strings"
	"fmt"
	"os"
)

const (
	LevelEmergency     = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInfo
	LevelDebug
)

const (
	AdapterConsole   = "console"
	AdapterFile      = "file"
)

const defaultAsyncMsgLen = 1000

type newLoggerFunc func() Logger

type Logger interface {
	Init(config string) error
	WriteMsg(when time.Time, msg string, level int) error
	Destroy()
	Flush()
}

type logMsg struct {
	level int
	msg   string
	when  time.Time
}

type Log struct {
	lock        sync.Mutex
	level       int
	msgChanLen  int64
	msgChan     chan *logMsg
	signalChan  chan string
	wg          sync.WaitGroup
	logger      Logger
	synchronous bool
}

var adapters    = make(map[string]newLoggerFunc)
var levelPrefix = [LevelDebug + 1]string{"[Emergency] ", "[Alert] ", "[Critical] ", "[Error] ", "[Warning] ", "[Notice] ", "[Info] ", "[Debug] "}

func Register(name string, Log newLoggerFunc) {
	if Log == nil {
		panic("Logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("Logs: Register called twice for provider " + name)
	}
	adapters[name] = Log
}

var logger = NewLogger()

func NewLogger() *Log{
	newLog := new(Log)
	newLog.level = LevelDebug
	newLog.msgChanLen = defaultAsyncMsgLen
	newLog.signalChan = make(chan string,1)
	return newLog
}

func (l *Log) setLogger(adapterName string,configs ...string) error {
	config := append(configs, "{}")[0]
	l.lock.Lock()
	defer l.lock.Unlock()

	log,ok := adapters[adapterName]
	if !ok{
		return fmt.Errorf("The adaptername(%s) can not found in this logger module",adapterName)
	}
	l.logger = log()
	err := l.logger.Init(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SetLogger: "+ err.Error())
		return fmt.Errorf("setLogger error :%s",err)
	}
	return nil
}

func SetLogger(adapterName string,configs ...string) error {
	return logger.setLogger(adapterName,configs...)
}

func (l *Log) setLevel(level int) {
	l.level = level
}

func SetLevel(level int) {
	logger.setLevel(level)
}

func (l *Log) Async( chanLen int64 ) *Log {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.synchronous {
		return l
	}
	l.synchronous = true

	if chanLen > 0 {
		l.msgChanLen = chanLen
	}
	l.msgChan = make(chan *logMsg,l.msgChanLen)
	l.wg.Add(1)
	go l.startLogger()
	return l
}

func Async(chanlen int64) *Log {
	return logger.Async(chanlen)
}

func (l *Log) writeMsg(level int,msg string,v ...interface{}){
	if len(v) > 0 {
		msg = fmt.Sprintf(msg,v...)
	}

	prefix := levelPrefix[level]
	msg    = prefix + msg
	when   := time.Now()

	if l.synchronous {
		l.msgChan <- &logMsg{level:level,msg:msg,when:when}
	}else{
		l.logger.WriteMsg(when,msg,level)
	}
}

func (l *Log) startLogger() {
	for{
		select {
		case lm := <- l.msgChan:
			l.logger.WriteMsg(lm.when,lm.msg,lm.level)
		case sg := <-l.signalChan:
			l.flush()
			if sg == "close" {
				l.logger.Destroy()
			}
			l.wg.Done()
			return
		}
	}
}

func (l *Log) Close() {
	if l.synchronous {
		l.signalChan <- "close"
		l.wg.Wait()
		close(l.msgChan)
	}else{
		l.logger.Flush()
		l.logger.Destroy()
	}
	close(l.signalChan)
}

func (l *Log) flush() {
	if l.synchronous {
		for{
			if len(l.msgChan) >0 {
				lm := <-l.msgChan
				msg   := lm.msg
				when  := lm.when
				level := lm.level
				l.logger.WriteMsg(when,msg,level)
			}else{
				break
			}
		}
	}
	l.logger.Flush()
}

func (l *Log) Flush() {
	if l.synchronous {
		l.signalChan <- "flush"
		l.wg.Wait()
		l.wg.Add(1)
		return
	}
	l.flush()
}

// Emergency Log EMERGENCY level message.
func (l *Log) Emergency(format string, v ...interface{}) {
	if LevelEmergency > l.level {
		return
	}
	l.writeMsg(LevelEmergency, format, v...)
}

// Alert Log ALERT level message.
func (l *Log) Alert(format string, v ...interface{}) {
	if LevelAlert > l.level {
		return
	}
	l.writeMsg(LevelAlert, format, v...)
}

// Critical Log CRITICAL level message.
func (l *Log) Critical(format string, v ...interface{}) {
	if LevelCritical > l.level {
		return
	}
	l.writeMsg(LevelCritical, format, v...)
}

// Error Log ERROR level message.
func (l *Log) Error(format string, v ...interface{}) {
	if LevelError > l.level {
		return
	}
	l.writeMsg(LevelError, format, v...)
}

// Warning Log WARNING level message.
func (l *Log) Warning(format string, v ...interface{}) {
	if LevelWarning > l.level {
		return
	}
	l.writeMsg(LevelWarning, format, v...)
}

// Notice Log NOTICE level message.
func (l *Log) Notice(format string, v ...interface{}) {
	if LevelNotice > l.level {
		return
	}
	l.writeMsg(LevelNotice, format, v...)
}

// Informational Log INFORMATIONAL level message.
func (l *Log) Informational(format string, v ...interface{}) {
	if LevelInfo > l.level {
		return
	}
	l.writeMsg(LevelInfo, format, v...)
}

// Debug Log DEBUG level message.
func (l *Log) Debug(format string, v ...interface{}) {
	if LevelDebug > l.level {
		fmt.Printf("LevelDebug:%d,l.level:%d\n",LevelDebug,l.level)
		return
	}
	l.writeMsg(LevelDebug, format, v...)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (l *Log) Warn(format string, v ...interface{}) {
	if LevelInfo > l.level {
		return
	}
	l.writeMsg(LevelInfo, format, v...)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (l *Log) Info(format string, v ...interface{}) {
	if LevelInfo > l.level {
		return
	}
	l.writeMsg(LevelInfo, format, v...)
}

// Emergency logs a message at emergency level.
func Emergency(f interface{}, v ...interface{}) {
	logger.Emergency(formatLog(f, v...))
}

// Alert logs a message at alert level.
func Alert(f interface{}, v ...interface{}) {
	logger.Alert(formatLog(f, v...))
}

// Critical logs a message at critical level.
func Critical(f interface{}, v ...interface{}) {
	logger.Critical(formatLog(f, v...))
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	logger.Error(formatLog(f, v...))
}

// Warning logs a message at warning level.
func Warning(f interface{}, v ...interface{}) {
	logger.Warning(formatLog(f, v...))
}

// Warn compatibility alias for Warning()
func Warn(f interface{}, v ...interface{}) {
	logger.Warn(formatLog(f, v...))
}

// Notice logs a message at notice level.
func Notice(f interface{}, v ...interface{}) {
	logger.Notice(formatLog(f, v...))
}

// Informational logs a message at info level.
func Informational(f interface{}, v ...interface{}) {
	logger.Info(formatLog(f, v...))
}

// Info compatibility alias for Warning()
func Info(f interface{}, v ...interface{}) {
	logger.Info(formatLog(f, v...))
}

// Debug logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	logger.Debug(formatLog(f, v...))
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
