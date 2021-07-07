package glog

import (
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
	LogPass // пропуск
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
	out io.Writer
}

var lpool = sync.Pool{
	New: func() interface{} {
		p := &pp{
			out: os.Stderr,
		}
		if pt, ok := os.LookupEnv("LOG_PATH"); ok {
			os.Mkdir(pt, os.ModePerm)
			if ex, err := os.Executable(); err == nil {
				p.out, err = os.OpenFile(path.Join(os.Getenv("LOG_PATH"), path.Base(ex)+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Fprint(os.Stderr, "logger pool err ", err)
				}
			}
		}
		return p
	},
}

func newPrinter(lv lvl) *pp {
	p := lpool.Get().(*pp)
	switch lv {
	case LogCrit:
		fmt.Fprint(p.out, "\033[35mC ")
	case LogErr:
		fmt.Fprint(p.out, "\033[31mE ")
	case LogWarn:
		fmt.Fprint(p.out, "\033[33mW ")
	case LogNote:
		fmt.Fprint(p.out, "\033[32mN ")
	case LogInf:
		fmt.Fprint(p.out, "\033[37mI ")
	case LogDeb:
		fmt.Fprint(p.out, "\033[36mD ")
	default:
		return p
	}
	fmt.Fprintf(p.out, "%-23s ", time.Now().In(loc).Format("2006-01-02 15:04:05.999"))
	_, fl, line, _ := runtime.Caller(3)
	fmt.Fprintf(p.out, "%s|%s:%d ▶ \033[0m ", path.Base(path.Dir(fl)), path.Base(fl), line)
	return p
}

func (p *pp) free() {
	lpool.Put(p)
}

func write(lv lvl, a ...interface{}) {
	if minLvl >= lv {
		p := newPrinter(lv)
		fmt.Fprintln(p.out, a...)
		p.free()
	}
}

func writeFormat(lv lvl, format string, a ...interface{}) {
	if minLvl >= lv {
		p := newPrinter(lv)
		if len(format) == 0 || format[len(format)-1] != '\n' {
			format += "\n"
		}
		fmt.Fprintf(p.out, format, a...)
		p.free()
	}
}

// Debug -
func Debug(a ...interface{}) {
	write(LogDeb, a...)
}

// Debugf -
func Debugf(format string, a ...interface{}) {
	writeFormat(LogDeb, format, a...)
}

// Info -
func Info(a ...interface{}) {
	write(LogInf, a...)
}

// Infof -
func Infof(format string, a ...interface{}) {
	writeFormat(LogInf, format, a...)
}

// Notice -
func Notice(a ...interface{}) {
	write(LogNote, a...)
}

// Noticef -
func Noticef(format string, a ...interface{}) {
	writeFormat(LogNote, format, a...)
}

// Warning -
func Warning(a ...interface{}) {
	write(LogWarn, a...)
}

// Warningf -
func Warningf(format string, a ...interface{}) {
	writeFormat(LogWarn, format, a...)
}

// Error -
func Error(a ...interface{}) {
	write(LogErr, a...)
}

// Errorf -
func Errorf(format string, a ...interface{}) {
	writeFormat(LogErr, format, a...)
}

// Fatal -
func Fatal(a ...interface{}) {
	write(LogCrit, a...)
	os.Exit(1)
}

// Fatalf -
func Fatalf(format string, a ...interface{}) {
	writeFormat(LogCrit, format, a...)
	os.Exit(1)
}

// Critical -
func Critical(a ...interface{}) {
	write(LogCrit, a...)
}

// Criticalf -
func Criticalf(format string, a ...interface{}) {
	writeFormat(LogCrit, format, a...)
}

// Writer returns the output destination for the standard logger.
// TO DO Delete
func Writer() io.Writer {
	return newPrinter(LogPass).out
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
