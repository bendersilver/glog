# glog

```sh
# .env file
LOG_PATH=path to log.			Default os.Stderr
LOG_TG=telegram bot token.		Default not send
LOG_TGIDS=user ids telegram.	Default not send	# LOG_TGIDS=112312,123213,124423144
LOG_LEVEL=log level.			Default DEBUG		# DEBUG || INFO || NOTICE || WARNING || ERROR
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