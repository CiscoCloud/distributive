package checks

import (
	"time"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/samuel/go-zookeeper/zk"
)

/*
#### ZooKeeperRUOK
Description: Are these Zookeeper servers responding to "ruok" requests?
Parameters:
- Timeout (time.Duration): Timeout for server response
- Servers ([]string): List of zookeeper servers
Example parameters:
- "5s", "20ms", "2h"
- "localhost:2181", "zookeeper.service.consul:2181"
*/
type ZooKeeperRUOK struct {
	timeout time.Duration
	servers []string
}

func init() { 
    chkutil.Register("ZooKeeperRUOK", func() chkutil.Check {
        return &ZooKeeperRUOK{}
    })
}

func (chk ZooKeeperRUOK) New(params []string) (chkutil.Check, error) {
	if len(params) < 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	dur, err := time.ParseDuration(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "time.Duration"}
	}
	chk.timeout = dur
	chk.servers = params[1:]
	return chk, nil
}

func (chk ZooKeeperRUOK) Status() (int, string, error) {
	oks := zk.FLWRuok(chk.servers, chk.timeout)
	var failed string
	// match zookeeper servers with failures for error message
	for i, ok := range oks {
		if !ok {
			failed += chk.servers[i]
			if i != len(oks)-1 {
				failed += ","
			}
		}
	}
	if failed == "" {
		return errutil.Success()
	}
	return 1, "Failed: " + failed, nil
}
