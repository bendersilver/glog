package glog

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	hs "github.com/bendersilver/help/sync"
	tb "gopkg.in/tucnak/telebot.v2"
)

func newTelelog() *teleLog {
	var token, ids string
	var ok bool
	if token, ok = os.LookupEnv("LOG_TG"); !ok {
		return nil
	}
	if ids, ok = os.LookupEnv("LOG_TGIDS"); !ok {
		return nil
	}
	var t teleLog
	t.exec, _ = os.Executable()
	t.bot, _ = tb.NewBot(tb.Settings{
		Token:       token,
		ParseMode:   tb.ModeMarkdown,
		Synchronous: true,
	})
	t.timeout = time.Second * 10
	if sec, ok := os.LookupEnv("LOG_TG_TIMEOUT"); ok {
		i, err := strconv.Atoi(sec)
		if err == nil {
			t.timeout = time.Second * time.Duration(i)
		}
	}
	if t.bot == nil {
		return nil
	}
	for _, id := range strings.Split(ids, ",") {
		i, err := strconv.Atoi(strings.TrimSpace(id))
		if err == nil {
			t.ids = append(t.ids, i)
		}
	}
	if len(t.ids) == 0 {
		return nil
	}
	return &t
}

type teleLog struct {
	timeout time.Duration
	exec    string
	bot     *tb.Bot
	ids     []int
	buf     sync.Map
	oq      hs.OnceQueue
}

func (t *teleLog) send() {
	t.oq.Do(func() {
		time.Sleep(t.timeout)
		var msg bytes.Buffer
		t.buf.Range(func(k, _ interface{}) bool {
			msg.WriteString(strings.ReplaceAll(k.(string), " ▶ \033[0m", " # "))
			t.buf.Delete(k)
			return true
		})
		if msg.Len() > 0 {
			for _, id := range t.ids {
				t.bot.Send(&tb.User{ID: id}, fmt.Sprintf("```sh\n%s\n%s\n```", t.exec, msg.Bytes()))
			}
		}
	})
}

func (t *teleLog) setValue(b *bytes.Buffer) {
	if b.Len() > 31 {
		t.buf.Store(b.String()[31:], nil)
		go t.send()
	}
}
