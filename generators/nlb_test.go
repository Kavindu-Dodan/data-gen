package generators

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

// refer http example of https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-access-logs.html#access-log-entry-format
const upstreamNLB = `tls 2.0 2020-04-01T08:51:42 net/my-network-loadbalancer/c6e77e28c25b2234 g3d4b5e8bb8464cd 72.21.218.154:51341 172.100.100.185:443 5 2 98 246 - arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99 - ECDHE-RSA-AES128-SHA tlsv12 - my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com h2 h2 "h2","http/1.1" 2020-04-01T08:51:20`

func Test_buildNLB(t *testing.T) {
	t.Run("Validate AWS documented HTTP NLB line", func(t *testing.T) {
		line := buildNLBLogLine(nlbCustomizer{
			time:              "2020-04-01T08:51:42",
			name:              "net/my-network-loadbalancer/c6e77e28c25b2234",
			elbID:             "g3d4b5e8bb8464cd",
			clientIPPort:      "72.21.218.154:51341",
			destinationIPPort: "172.100.100.185:443",
			conMs:             5,
			tlsHSMs:           2,
			receivedBytes:     98,
			sentBytes:         246,
			tlsAert:           "-",
			certARN:           "arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99",
			cipher:            "ECDHE-RSA-AES128-SHA",
			protocol:          "tlsv12",
			domain:            "my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com",
			feProtocol:        "h2",
			beProtocol:        "h2",
			alpnList:          `"h2","http/1.1"`,
			creationTime:      "2020-04-01T08:51:20",
		})

		require.Equal(t, upstreamNLB, strings.TrimSpace(line))
	})
}
