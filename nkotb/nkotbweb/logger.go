package nkotbweb

import "log"

var logger = log.Default()

func init() {
	logger.SetPrefix(AppName + " ")
	logger.SetFlags(log.LstdFlags | log.Lshortfile)
}
