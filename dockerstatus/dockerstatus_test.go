package dockerstatus

import (
	"reflect"
	"testing"
)

var mantlTestStr = `CONTAINER ID        IMAGE                         COMMAND                  CREATED             STATUS              PORTS                                                                                              NAMES
173eb386ecd5        ciscocloud/logstash:1.5.3     "/docker-entrypoint.s"   11 minutes ago      Up 11 minutes       0.0.0.0:1514->1514/tcp, 0.0.0.0:8125->8125/udp, 0.0.0.0:25826->25826/udp, 0.0.0.0:5678->5678/tcp   logstash
fb617d0b0d12        ciscocloud/nginx-consul:1.2   "/scripts/launch.sh"     15 minutes ago      Up 15 minutes                                                                                                          nginx-consul
`

var emptyTestStr = `CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
`

func TestParseRunningContainers(t *testing.T) {
	t.Parallel()
	outputs := [][]string{
		{"ciscocloud/logstash:1.5.3", "ciscocloud/nginx-consul:1.2"}, {},
	}
	for i, input := range []string{mantlTestStr, emptyTestStr} {
		expected := outputs[i]
		actual := parseRunningContainers(input)
		if !reflect.DeepEqual(actual, expected) {
			t.Logf("Expected: %v\nActual: %v", expected, actual)
		}
	}
}
