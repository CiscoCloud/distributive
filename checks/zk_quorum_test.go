package checks

import (
    "testing"
//    "github.com/CiscoCloud/distributive/checks"
)

func TestZooKeeperQuorumConfig(t *testing.T) {
    chk := ZooKeeperQuorum{config: "fixtures/zoo_test.cfg"}
    servers, err := chk.LoadConfig()
    if err != nil {
        t.Errorf("error opening fixture: %s", err.Error())
    }
    if servers == nil {
        t.Error("return empty slice")
    } else {
        for i, val := range []string{
            "avnik-london-control-01:2888",
            "avnik-london-control-02:2888",
            "avnik-london-control-03:2888",
        } {
            if servers[i] != val {
                t.Errorf("got '%s' when expected '%s'", servers[i], val)
            }
        }
    }
}
