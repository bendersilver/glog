# glog

```sh
# .env file
LOG_PATH=/root/log/				Default os.Stderr
LOG_TG=<token>					Default not send
LOG_TGIDS=teleId,teleId			Default not send	# LOG_TGIDS=112312,123213,124423144
LOG_LEVEL=DEBUG					Default DEBUG		# DEBUG || INFO || NOTICE || WARNING || ERROR
LOG_TG_TIMEOUT=10				Default 10 sec
```

```go
package main

import (
	"time"

	"github.com/bendersilver/glog"
)

func main() {
	glog.Debug("Debug")
	glog.Info("Info")
	glog.Notice("Notice")
}
```