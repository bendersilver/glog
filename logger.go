package glog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
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
)

type pp struct {
	sync.Mutex
	out io.Writer
	buf *bytes.Buffer
}

var pool *pp

func init() {
	pool = &pp{
		buf: new(bytes.Buffer),
		out: os.Stdout,
	}
}

func (p *pp) writeHead(lv lvl) {
	switch lv {
	case LogCrit:
		// fmt.Fprint(p.buf, "\033[35mC ")
		fmt.Fprint(p.buf, "CRT ")
	case LogErr:
		// fmt.Fprint(p.buf, "\033[31mE ")
		fmt.Fprint(p.buf, "ERR ")
	case LogWarn:
		// fmt.Fprint(p.buf, "\033[33mW ")
		fmt.Fprint(p.buf, "WRN ")
	case LogNote:
		// fmt.Fprint(p.buf, "\033[32mN ")
		fmt.Fprint(p.buf, "NTC ")
	case LogInf:
		// fmt.Fprint(p.buf, "\033[37mI ")
		fmt.Fprint(p.buf, "INF ")
	case LogDeb:
		// fmt.Fprint(p.buf, "\033[36mD ")
		fmt.Fprint(p.buf, "DBG ")
	default:
		// fmt.Fprint(p.buf, "\033[37mI ")
		fmt.Fprint(p.buf, "INF ")
	}
	_, fl, line, _ := runtime.Caller(3)
	// fmt.Fprintf(p.buf, "%s:%s:%d ▶ \033[0m", path.Base(path.Dir(fl)), path.Base(fl), line)
	fmt.Fprintf(p.buf, "%s:%s:%d >> ", path.Base(path.Dir(fl)), path.Base(fl), line)
}

func (p *pp) free() {
	io.Copy(p.out, p.buf)
	p.buf.Truncate(0)
}

func (p *pp) write(lv lvl, a ...interface{}) {
	p.Lock()
	defer p.Unlock()
	p.writeHead(lv)
	fmt.Fprintln(p.buf, a...)
	p.free()
}

func (p *pp) writeFormat(lv lvl, format string, a ...interface{}) {
	p.Lock()
	defer p.Unlock()
	p.writeHead(lv)
	if len(format) == 0 || format[len(format)-1] != '\n' {
		format += "\n"
	}
	fmt.Fprintf(p.buf, format, a...)
	p.free()
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

// Recover -
func Recover(f func()) {
	defer func() {
		if r := recover(); r != nil {
			pool.writeFormat(LogCrit, "%v\n%s", r, debug.Stack())
		}
	}()
	f()
}
