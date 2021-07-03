package glog

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

type teleLog struct {
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
			fmt.Fprint(os.Stderr, "logger telegram pool err ", err)
			return t
		}
		t.bot, err = tb.NewBot(tb.Settings{
			Token:       os.Getenv("LOG_TG"),
			ParseMode:   tb.ModeMarkdown,
			Synchronous: true,
		})
		if err != nil {
			Error("init leleLog error: ", err)
		}
		if ids, ok := os.LookupEnv("LOG_TGIDS"); ok {
			for _, id := range strings.Split(ids, ",") {
				i, err := strconv.Atoi(strings.TrimSpace(id))
				if err != nil {
					Error("error parse id:", id, err)
				} else {
					t.ids = append(t.ids, i)
				}
			}
		} else {
			Error("not set LOG_TGIDS variable")
			Info("example: LOG_TGIDS=8889112,234234")
		}
		return t
	},
}

func newTeleLog() *teleLog {
	return tpool.Get().(*teleLog)
}

func (t *teleLog) free() {
	lpool.Put(t)
}

func send() {
	t := newTeleLog()
	defer t.free()
	if t.bot != nil {
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
			t.buf.Range(func(key, _ interface{}) bool {
				msg += fmt.Sprintf("%s\n", key)
				t.buf.Delete(key)
				return true
			})
			t.timer = nil
			go func() {
				for _, id := range t.ids {
					_, err := t.bot.Send(&tb.User{ID: id}, msg)
					if err != nil {
						Error(err)
					}
				}
			}()
		})
	}

}

// Tg -
func Tg(l lvl, a ...interface{}) {
	t := newTeleLog()
	t.buf.Store(fmt.Sprint(a...), nil)
	t.free()
	write(l, a...)
	send()
}

// Tgf -
func Tgf(l lvl, format string, a ...interface{}) {
	t := newTeleLog()
	t.buf.Store(fmt.Sprint(a...), nil)
	t.free()
	writeFormat(l, format, a...)
	send()
}
