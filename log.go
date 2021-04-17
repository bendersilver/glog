package glog

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type lvl int

// Log levels.
const (
	LogCrit lvl = iota
	LogErr
	LogWarn
	LogNote
	LogInf
	LogDeb
	LogPass // пропуск
)

var std = &logger{
	out:   os.Stderr,
	level: LogDeb,
}

// auto init path
func init() {
	var f *os.File
	var err error
	exe, _ := os.Executable()
	var pth = []string{"./", path.Dir(os.Args[0]), exe}
	for _, p := range pth {
		f, err = os.OpenFile(path.Join(p, ".env"), os.O_RDONLY, 0644)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		spl := strings.Split(scanner.Text(), "=")
		if len(spl) > 1 {
			switch spl[0] {
			case "LOG_PATH":
				SetPath(spl[1])
			case "LOG_TG":
				std.bot, err = tb.NewBot(tb.Settings{
					Token:       spl[1],
					ParseMode:   tb.ModeMarkdown,
					Synchronous: true,
				})
				if err != nil {
					Warning(err)
				}
			case "LOG_TG_ID":
				i, err := strconv.Atoi(spl[1])
				if err != nil {
					Warning(err)
				} else {
					std.botID = &tb.Chat{ID: int64(i)}
				}
			}
		}
	}
	if std.bot != nil && std.botID != nil {
		go tgLoop()
		std.botBuf = new(bytes.Buffer)
	}
}

func tgLoop() {
	var str string
	var err error
	tik := time.Tick(time.Second * 2)
	for range tik {
		if std.botBuf.Len() > 0 {
			std.mu.Lock()
			str = std.botBuf.String()
			std.botBuf.Reset()
			std.mu.Unlock()
			_, err = std.bot.Send(std.botID, fmt.Sprintf("```\n%s\n```", str))
			if err != nil {
				Error(err)
			}
		}
	}
}

type logger struct {
	mu     sync.Mutex
	out    io.Writer
	buf    []byte
	level  lvl
	bot    *tb.Bot
	botID  *tb.Chat
	botBuf *bytes.Buffer
}

// Writer returns the output destination for the logger.
func (l *logger) Writer() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

func (l *logger) Output(lv lvl, s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = l.buf[:0]
	switch lv {
	case LogCrit:
		l.buf = append(l.buf, "\033[35mC "...)
	case LogErr:
		l.buf = append(l.buf, "\033[31mE "...)
	case LogWarn:
		l.buf = append(l.buf, "\033[33mW "...)
	case LogNote:
		l.buf = append(l.buf, "\033[32mN "...)
	case LogInf:
		l.buf = append(l.buf, "\033[37mI "...)
	case LogDeb:
		l.buf = append(l.buf, "\033[36mD "...)
	}
	_, fn, line, _ := runtime.Caller(2)
	l.buf = append(l.buf, fmt.Sprintf("%-23s %s:%d ▶ \033[0m %s",
		time.Now().Format("2006-01-02 15:04:05.999"),
		path.Base(fn),
		line,
		s,
	)...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

// Writer returns the output destination for the standard logger.
func Writer() io.Writer {
	return std.Writer()
}

func tgLog(v ...interface{}) {
	if std.botBuf != nil {
		std.mu.Lock()
		if std.botBuf.Len() < 1024 {
			_, fn, line, _ := runtime.Caller(2)
			std.botBuf.WriteString(fmt.Sprintf("%-12s %s:%d # %s",
				time.Now().Format("15:04:05.999"),
				path.Base(fn),
				line,
				fmt.Sprintln(v...),
			))
		}
		std.mu.Unlock()
	} else {
		Warning("set env variable LOG_TG and LOG_TG_ID")
	}
}

// Logger -
func Logger(tag string) *log.Logger {
	return log.New(std.out, tag, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

// SendTg -
func SendTg(l lvl, v ...interface{}) {
	tgLog(v...)
	if l != LogPass {
		std.Output(l, fmt.Sprintln(v...))
	}

}

// SendTgf -
func SendTgf(l lvl, format string, a ...interface{}) {
	SendTg(l, fmt.Sprintf(format, a...))
}

// Debug -
func Debug(v ...interface{}) {
	std.Output(LogDeb, fmt.Sprintln(v...))
}

// Debugf -
func Debugf(format string, a ...interface{}) {
	std.Output(LogDeb, fmt.Sprintf(format, a...))
}

func Info(v ...interface{}) {
	std.Output(LogInf, fmt.Sprintln(v...))
}

// Infof -
func Infof(format string, a ...interface{}) {
	std.Output(LogInf, fmt.Sprintf(format, a...))
}

// Notice -
func Notice(v ...interface{}) {
	std.Output(LogNote, fmt.Sprintln(v...))
}

// Noticef -
func Noticef(format string, a ...interface{}) {
	std.Output(LogNote, fmt.Sprintf(format, a...))
}

// Warning -
func Warning(v ...interface{}) {
	std.Output(LogWarn, fmt.Sprintln(v...))
}

// Warningf -
func Warningf(format string, a ...interface{}) {
	std.Output(LogWarn, fmt.Sprintf(format, a...))
}

// Error -
func Error(v ...interface{}) {
	std.Output(LogErr, fmt.Sprintln(v...))
}

// Errorf -
func Errorf(format string, a ...interface{}) {
	std.Output(LogErr, fmt.Sprintf(format, a...))
}

// Fatal -
func Fatal(v ...interface{}) {
	std.Output(LogCrit, fmt.Sprintln(v...))
	os.Exit(1)
}

// Fatalf -
func Fatalf(format string, a ...interface{}) {
	std.Output(LogCrit, fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Critical -
func Critical(v ...interface{}) {
	std.Output(LogCrit, fmt.Sprintln(v...))
}

// Criticalf -
func Criticalf(format string, a ...interface{}) {
	std.Output(LogCrit, fmt.Sprintf(format, a...))
}

// SetPath - set path loller file
func SetPath(p string) {
	os.Mkdir(p, os.ModePerm)
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	flName := strings.Join([]string{path.Base(ex), "log"}, ".")
	std.out, err = os.OpenFile(path.Join(p, flName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}
