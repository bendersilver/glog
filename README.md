# glog

```sh
# .env file
LOG_PATH=path to log
LOG_TG=telegram bot token
LOG_TGIDS=user ids telegram # LOG_TGIDS=112312,123213,124423144
```

```go
package main

import (
	"time"

	"github.com/bendersilver/glog"
)

func main() {
	glog.Tg(glog.LogErr, "sending to telegrams once every 5 seconds")
	glog.Tg(glog.LogPass, "sending to telegrams. no write log")
	time.Sleep(time.Second * 6)
	glog.Debug("Debug")
	glog.Info("Info")
	glog.Notice("Notice")
}
```