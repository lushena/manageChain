package datadump

import "time"

// constans for blockfile dump
const (
	DumpForCronTab        = 0
	DumpForBlockFileLimit = 1
)

var DumpReasonIndex_name = map[int]string{
	0: "CRONTAB",
	1: "BLOCKFILELIMIT",
}

// config for blockfile dump
type DumpConf struct {
	Enabled        bool
	DumpDir        string
	LoadDir        string
	MaxFileLimit   int
	DumpCron       []string
	DumpInterval   time.Duration
	LoadRetryTimes int
}
