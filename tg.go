package glog

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

type teleLog struct {
	exec  string
	bot   *tb.Bot
	ids   []int
	buf   sync.Map
	timer *time.Timer
}

var tpool = sync.Pool{
	New: func() interface{} {
		var err error
		t := new(teleLog)
		err = godotenv.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "logger telegram pool err: %v\n", err)
			return t
		}
		if tk, ok := os.LookupEnv("LOG_TG"); ok {
			t.exec, _ = os.Executable()
			t.bot, err = tb.NewBot(tb.Settings{
				Token:       tk,
				ParseMode:   tb.ModeMarkdown,
				Synchronous: true,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "init telebot error: %v\n", err)
			}
		} else {
			fmt.Fprint(os.Stderr, "not set LOG_TG variable\n")
			return t
		}

		if ids, ok := os.LookupEnv("LOG_TGIDS"); ok {
			for _, id := range strings.Split(ids, ",") {
				i, err := strconv.Atoi(strings.TrimSpace(id))
				if err != nil {
					fmt.Fprintf(os.Stderr, "error parse id: ID: %s, err: %v\n", id, err)
				} else {
					t.ids = append(t.ids, i)
				}
			}
		} else {
			fmt.Fprint(os.Stderr, "not set LOG_TGIDS variable\n")
		}
		return t
	},
}

func newTeleLog() *teleLog {
	return tpool.Get().(*teleLog)
}

func (t *teleLog) free() {
	tpool.Put(t)
}

func send() {
	t := newTeleLog()
	defer t.free()
	if t.bot == nil {
		Warning("teleLog not set")
		return
	}
	if len(t.ids) == 0 {
		Warning("teleLog not set users recipient")
		return
	}
	if t.timer == nil {
		t.timer = time.AfterFunc(time.Second*5, func() {
			var msg string
			t.buf.Range(func(k, v interface{}) bool {
				msg += fmt.Sprintf("%s  #  %s\n", v, k)
				t.buf.Delete(k)
				return true
			})
			t.timer = nil
			go func(name string) {
				for _, id := range t.ids {
					_, err := t.bot.Send(&tb.User{ID: id}, fmt.Sprintf("```sh\n%s\n%s\n```", name, msg))
					if err != nil {
						Error(err)
					}
				}
			}(t.exec)
		})
	}

}

// Tg -
func Tg(l lvl, a ...interface{}) {
	t := newTeleLog()
	_, fl, line, _ := runtime.Caller(3)
	t.buf.Store(fmt.Sprint(a...), fmt.Sprintf("%s:%d", path.Base(fl), line))
	t.free()
	write(l, a...)
	send()
}

// Tgf -
func Tgf(l lvl, format string, a ...interface{}) {
	t := newTeleLog()
	_, fl, line, _ := runtime.Caller(3)
	t.buf.Store(fmt.Sprintf(format, a...), fmt.Sprintf("%s:%d", path.Base(fl), line))
	t.free()
	writeFormat(l, format, a...)
	send()
}
