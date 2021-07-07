package glog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type lvl int

// Log levels.
const (
	LogCrit lvl = 1 << iota
	LogErr
	LogWarn
	LogNote
	LogInf
	LogDeb
	LogPass // pass log
)

var loc *time.Location
var minLvl lvl

func init() {
	minLvl = LogDeb
	loc, _ = time.LoadLocation("Asia/Yekaterinburg")
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "logger pool err: %v\n", err)
	}
	if lv, ok := os.LookupEnv("LOG_LEVEL"); ok {
		switch lv {
		case "DEBUG":
			minLvl = LogDeb
		case "INFO":
			minLvl = LogInf
		case "NOTICE":
			minLvl = LogNote
		case "WARNING":
			minLvl = LogWarn
		case "ERROR":
			minLvl = LogErr
		default:
			fmt.Fprintf(os.Stderr, "log level: %s not found\n", lv)
		}
	}
}

// SetTimeZone -
func SetTimeZone(tz string) (err error) {
	loc, err = time.LoadLocation(tz)
	return
}

type pp struct {
	sync.Mutex
	out  io.Writer
	buf  *bytes.Buffer
	tele *teleLog
}

var pool *pp

func init() {
	godotenv.Load()
	pool = &pp{
		buf:  new(bytes.Buffer),
		out:  os.Stdout,
		tele: newTelelog(),
	}
	if pt, ok := os.LookupEnv("LOG_PATH"); ok {
		os.Mkdir(pt, os.ModePerm)
		if ex, err := os.Executable(); err == nil {
			pool.out, err = os.OpenFile(path.Join(os.Getenv("LOG_PATH"), path.Base(ex)+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprint(os.Stdout, "logger pool err ", err)
			}
		}
	}
}

func (p *pp) writeHead(lv lvl) {
	switch lv {
	case LogCrit:
		fmt.Fprint(p.buf, "\033[35mC ")
	case LogErr:
		fmt.Fprint(p.buf, "\033[31mE ")
	case LogWarn:
		fmt.Fprint(p.buf, "\033[33mW ")
	case LogNote:
		fmt.Fprint(p.buf, "\033[32mN ")
	case LogInf:
		fmt.Fprint(p.buf, "\033[37mI ")
	case LogDeb:
		fmt.Fprint(p.buf, "\033[36mD ")
	default:
		fmt.Fprint(p.buf, "\033[37mI ")
	}
	_, fl, line, _ := runtime.Caller(3)
	fmt.Fprintf(p.buf, "%-23s %s:%s:%d ▶ \033[0m", time.Now().In(loc).Format("2006-01-02 15:04:05.999"), path.Base(path.Dir(fl)), path.Base(fl), line)
}

func (p *pp) free() {
	if p.tele != nil {
		p.tele.setValue(p.buf)
	}
	io.Copy(p.out, p.buf)
	p.buf.Truncate(0)
}

func (p *pp) write(lv lvl, a ...interface{}) {
	if minLvl >= lv {
		p.Lock()
		defer p.Unlock()
		p.writeHead(lv)
		fmt.Fprintln(p.buf, a...)
		p.free()
	}
}

func (p *pp) writeFormat(lv lvl, format string, a ...interface{}) {
	if minLvl >= lv {
		p.Lock()
		defer p.Unlock()
		p.writeHead(lv)
		if len(format) == 0 || format[len(format)-1] != '\n' {
			format += "\n"
		}
		fmt.Fprintf(p.buf, format, a...)
		p.free()
	}
}

// Debug -
func Debug(a ...interface{}) {
	pool.write(LogDeb, a...)
}

// Debugf -
func Debugf(format string, a ...interface{}) {
	pool.writeFormat(LogDeb, format, a...)
}

// Info -
func Info(a ...interface{}) {
	pool.write(LogInf, a...)
}

// Infof -
func Infof(format string, a ...interface{}) {
	pool.writeFormat(LogInf, format, a...)
}

// Notice -
func Notice(a ...interface{}) {
	pool.write(LogNote, a...)
}

// Noticef -
func Noticef(format string, a ...interface{}) {
	pool.writeFormat(LogNote, format, a...)
}

// Warning -
func Warning(a ...interface{}) {
	pool.write(LogWarn, a...)
}

// Warningf -
func Warningf(format string, a ...interface{}) {
	pool.writeFormat(LogWarn, format, a...)
}

// Error -
func Error(a ...interface{}) {
	pool.write(LogErr, a...)
}

// Errorf -
func Errorf(format string, a ...interface{}) {
	pool.writeFormat(LogErr, format, a...)
}

// Fatal -
func Fatal(a ...interface{}) {
	pool.write(LogCrit, a...)
	os.Exit(1)
}

// Fatalf -
func Fatalf(format string, a ...interface{}) {
	pool.writeFormat(LogCrit, format, a...)
	os.Exit(1)
}

// Critical -
func Critical(a ...interface{}) {
	pool.write(LogCrit, a...)
}

// Criticalf -
func Criticalf(format string, a ...interface{}) {
	pool.writeFormat(LogCrit, format, a...)
}

// Writer returns the output destination for the standard logger.
// TO DO Delete
func Writer() io.Writer {
	return pool.out
}

// Logger -
func Logger(tag string) *log.Logger {
	var out = os.Stderr
	if pt, ok := os.LookupEnv("LOG_PATH"); ok {
		os.Mkdir(pt, os.ModePerm)
		if ex, err := os.Executable(); err == nil {
			f, err := os.OpenFile(path.Join(ex, path.Base(ex)+fmt.Sprintf(".logger.%s.log", tag)), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				Error("Logger", err)
			} else {
				out = f
			}
		}
	}
	return log.New(out, tag, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}
