package checks

import (
	"errors"
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

func (chk ZooKeeperRUOK) ID() string { return "ZooKeeperRUOK" }

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

/*
#### ZooKeeperLatency
Description: Is the latency of these zookeeper servers below this value?
Parameters:
- Max latency (time.Duration): Maximum acceptable average latency
- Servers ([]string): List of zookeeper servers
Example parameters:
- "5s", "20ms", "2h"
- "localhost:2181", "zookeeper.service.consul:2181"
*/
type ZooKeeperLatency struct {
	maxLatency time.Duration
	servers    []string
}

func (chk ZooKeeperLatency) ID() string { return "ZooKeeperLatency" }

func (chk ZooKeeperLatency) New(params []string) (chkutil.Check, error) {
	if len(params) < 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	dur, err := time.ParseDuration(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "time.Duration"}
	}
	chk.maxLatency = dur
	chk.servers = params[1:]
	return chk, nil
}

func (chk ZooKeeperLatency) Status() (int, string, error) {
	stats, ok := zk.FLWSrvr(chk.servers, chk.maxLatency)
	if !ok {
		if len(stats) > 1 {
			for _, srv := range stats {
				if srv.Error != nil {
					return 1, "", srv.Error
				}
			}
		}
		return 1, "", errors.New("Unknown error in zk api")
	}
	var failed string
	// match zookeeper servers with failures for error message
	for i, stat := range stats {
		if stat.AvgLatency > int64(chk.maxLatency) {
			failed += chk.servers[i]
		}
	}
	if failed == "" {
		return errutil.Success()
	}
	return 1, "Latency too high: " + failed, nil
}
