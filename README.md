# glog

```sh
# .env file
LOG_PATH=path to log
LOG_TG=telegram bot token
LOG_TG_ID=user id telegram
```

```go
package main

import (
	"time"

	"github.com/bendersilver/glog"
)

// .env
// LOG_PATH=path to log
// LOG_TG=telegram bot token
// LOG_TG_ID=user id telegram

func main() {
	glog.Tg("sending to telegrams once every 10 seconds")
	time.Sleep(time.Second * 11)
	glog.Debug("Debug")
	glog.Info("Info")
	glog.Notice("Notice")
	// ....
	// logger http server
	// logger = glog.Logger("prfix logger ")
}
```