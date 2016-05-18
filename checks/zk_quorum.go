package checks

import (
    "bufio"
    "fmt"
	"time"
    "strings"
    "regexp"
    "os"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/samuel/go-zookeeper/zk"
)

/*
#### ZooKeeperQuorum
Description: Are these Zookeeper servers responding to "ruok" requests?
Parameters:
- Timeout (time.Duration): Timeout for server response
- Config file: file with zk config, where all nodes listed
Example parameters:
- "5s", "20ms", "2h"
- "/etc/zookeeper/conf/zoo.cfg"
*/
type ZooKeeperQuorum struct {
	timeout time.Duration
    config  string
}

func init() { 
    chkutil.Register("ZooKeeperQuorum", func() chkutil.Check {
        return &ZooKeeperQuorum{
            config: "/etc/zookeeper/conf/zoo.cfg",  /* default value, only for new k/v arg parsing */
        }
    })
}

func (chk ZooKeeperQuorum) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	dur, err := time.ParseDuration(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "time.Duration"}
	}
	chk.timeout = dur
	chk.config = params[1]
	return chk, nil
}


func (chk ZooKeeperQuorum) LoadConfig() ([]string, error) {
    beginOfServerLine := regexp.MustCompile("^server\\.\\d+=(.*)$")
    file, err := os.Open(chk.config)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var servers []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if match := beginOfServerLine.FindStringSubmatch(line); match != nil {
            parts := strings.Split(match[1], ":")
            host := parts[0]
            port := parts[1]
            servers = append(servers, fmt.Sprintf("%s:%s", host, port))
        }
    }
    return servers, nil
}

func (chk ZooKeeperQuorum) Status() (int, string, error) {
    servers, err := chk.LoadConfig()
    if err != nil {
        return 1, "Failed: " + err.Error(), nil
    }

	oks, _ := zk.FLWSrvr(servers, chk.timeout)

    var leaders int = 0
    var followers int = 0
    var unknowns int = 0

	var failed []string

    // Single standalone server
    if len(servers) == 0  && len(oks) == 1 {
        ok := oks[0]
        if ok == nil {
            return 1, fmt.Sprintf("%s: failed to connect", servers[0]), nil
        }
        if ok.Mode != zk.ModeStandalone {
            return 1, fmt.Sprintf("%s: mode is '%s', when should be 'standalone' for single server", servers[0], ok.Mode), nil
        }
		return errutil.Success()
    }

    for _, ok := range oks {
        if ok != nil {
            switch ok.Mode {
            case zk.ModeLeader:
                leaders++
                break
            case zk.ModeFollower:
                followers++
                break
            default:
                unknowns++
            }
        } else {
            unknowns++
        }
    }

    if leaders != 1 {
        failed = append(failed, fmt.Sprintf("Cluster have more than one leader"))
    }

    // FIXME: add more checks here


	if len(failed) == 0 {
		return errutil.Success()
	}
	return 1, "Failed: " + strings.Join(failed, ", ") , nil
}
