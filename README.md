# glog

```sh
# .env file
LOG_PATH=/root/log/				Default os.Stderr
LOG_LEVEL=DEBUG					Default DEBUG		# DEBUG || INFO || NOTICE || WARNING || ERROR
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