package main

import (
	"github.com/gao88/autoupdate"
	"time"
)

func main() {
	autoupdate.PushCount(10) //在线有10人

	time.Sleep(8 * time.Second)

	autoupdate.PushCount(0) //过8秒后有在线有0人

	time.Sleep(15 * time.Second)
}
