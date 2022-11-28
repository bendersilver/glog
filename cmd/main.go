package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bendersilver/glog"
	"github.com/imroc/req"
)

// journalctl --user -n 10 -f -o cat

type buf struct {
	sync.Mutex
	buf      bytes.Buffer
	lastUnit string
}

// id -u bot
func main() {
	if os.Getenv("BOT") == "" {
		glog.Error("not set env var BOT")
	}
	if os.Getenv("IDS") == "" {
		glog.Error("not set env var IDS")
	}

	cmd := exec.Command("journalctl", "--user", "-f", "-o", "json")
	r, err := cmd.StdoutPipe()
	if err != nil {
		glog.Fatal(err)
	}

	var b buf
	go scan(bufio.NewScanner(r), &b)
	go sender(&b)

	err = cmd.Start()
	if err != nil {
		glog.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		glog.Fatal(err)
	}
}

func scan(s *bufio.Scanner, b *buf) {

	var data struct {
		Exe  string          `json:"_EXE"`
		Msg  json.RawMessage `json:"MESSAGE"`
		Com  string          `json:"_COMM"`
		Unit string          `json:"_SYSTEMD_UNIT"`
	}
	var err error
	for s.Scan() {
		err = json.Unmarshal(s.Bytes(), &data)
		if err != nil {
			glog.Error(err, s.Text())
		} else {
			if strings.HasPrefix(data.Unit, os.Getenv("UNIT_PERFIX")) {
				b.Lock()
				if b.lastUnit != data.Unit {
					b.buf.WriteString(data.Exe + "\n")
				}
				b.lastUnit = data.Unit
				var msg string
				err = json.Unmarshal(data.Msg, &msg)
				if err == nil {
					b.buf.WriteString(msg)
				} else {
					var bt []byte
					json.Unmarshal(data.Msg, &bt)
					b.buf.Write(bt)
				}
				b.buf.WriteString("\n")
				b.Unlock()
			}
		}
	}
}

func sender(b *buf) {
	var resp struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"error_code"`
		Desc    string `json:"description"`
	}

	var data struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
		Mode   string `json:"parse_mode"`
	}
	data.Mode = "Markdown"

	var users []int64
	for _, v := range strings.Split(os.Getenv("IDS"), ",") {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			glog.Error(err)
		} else {
			users = append(users, id)
		}
	}

	for {
		if b.buf.Len() > 0 {
			b.Lock()
			data.Text = fmt.Sprintf("```\n%s\n```", strings.TrimSpace(b.buf.String()))
			b.buf.Reset()
			b.lastUnit = ""
			b.Unlock()
			for _, id := range users {
				data.ChatID = id
				r, err := req.Post("https://api.telegram.org/bot"+os.Getenv("BOT")+"/sendMessage",
					req.BodyJSON(data),
				)
				if err != nil {
					glog.Error(err)
					continue
				}
				err = r.ToJSON(&resp)
				if err != nil {
					glog.Error(err)
					continue
				}
				if !resp.OK {
					glog.Error(resp.ErrCode, resp.Desc)
				}
			}
			if users == nil {
				glog.Debug(data.Text)
			}
		}
		time.Sleep(time.Second)
	}
}
