package checks

import (
    "fmt"
	"time"
    "strconv"
    "strings"

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
    chkutil.Register("ServerStats", func() chkutil.Check {
        return &ZooKeeperServerStats{}
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

type ZooKeeperServerStats struct {
	timeout time.Duration
	servers []string
	minLatency int64
	maxLatency int64
	avgLatency int64
}


func (chk ZooKeeperServerStats) New(params []string) (chkutil.Check, error) {
	if len(params) < 5 {
		return chk, errutil.ParameterLengthError{5, params}
	}
	dur, err := time.ParseDuration(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "time.Duration"}
	}

    chk.minLatency, err = strconv.ParseInt(params[1], 0, 64)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "minLatency"}
	}

    chk.avgLatency, err = strconv.ParseInt(params[2], 0, 64)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[2], "avgLatency"}
	}

    chk.maxLatency, err = strconv.ParseInt(params[3], 0, 64)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[3], "maxLatency"}
	}

	chk.timeout = dur
	chk.servers = params[4:]
	return chk, nil
}

func (chk ZooKeeperServerStats) Status() (int, string, error) {
	oks, _ := zk.FLWSrvr(chk.servers, chk.timeout)
	var failed []string
	// match zookeeper servers with failures for error message
	for i, ok := range oks {
		if ok == nil {
            failed = append(failed, fmt.Sprintf("%s: failed to connect", chk.servers[i]))
        }
        if ok.MinLatency > chk.minLatency {
            failed = append(failed, fmt.Sprintf("%s: min latency too big: %v", chk.servers[i], ok.MinLatency)) 
        }
        if ok.MaxLatency > chk.maxLatency {
            failed = append(failed, fmt.Sprintf("%s: max latency too big: %v", chk.servers[i], ok.MaxLatency)) 
        }
        if ok.AvgLatency > chk.avgLatency {
            failed = append(failed, fmt.Sprintf("%s: avg latency too big: %v", chk.servers[i], ok.AvgLatency)) 
        }
    }
	if len(failed) == 0 {
		return errutil.Success()
	}
	return 1, "Failed: " + strings.Join(failed, ", ") , nil
}
