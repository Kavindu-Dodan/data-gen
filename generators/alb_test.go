package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// refer http example of https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html#access-log-entry-format
const upstreamALBHTTP = `http 2018-07-02T22:23:00.186641Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.000 0.001 0.000 200 200 34 366 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.46.0" - - arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337262-36d228ad5d99923122bbe354" "-" "-" 0 2018-07-02T22:22:48.364000Z "forward" "-" "-" "10.0.0.1:80" "200" "-" "-" TID_1234abcd5678ef90`

func Test_buildALB(t *testing.T) {
	t.Run("Validate AWS documented HTTP ALB line", func(t *testing.T) {
		line := buildALBLogLine(albCustomizer{
			logType:          "http",
			timestamp:        "2018-07-02T22:23:00.186641Z",
			creationTime:     "2018-07-02T22:22:48.364000Z",
			request:          `"GET http://www.example.com:80/ HTTP/1.1"`,
			elbID:            "app/my-loadbalancer/50dc6c495c0c9188",
			targetARN:        "arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067",
			traceID:          `"Root=1-58337262-36d228ad5d99923122bbe354"`,
			userAgent:        `"curl/7.46.0"`,
			clientIPPort:     "192.168.131.39:2817",
			targetIPPort:     "10.0.0.1:80",
			targetPortList:   `"10.0.0.1:80"`,
			requestID:        "TID_1234abcd5678ef90",
			requestProcTime:  0.000,
			targetProcTime:   0.001,
			responseProcTime: 0.000,
			elbStatus:        "200",
			targetStatus:     "200",
			cipher:           "-",
			sslProtocol:      "-",
			domain:           `"-"`,
			receivedBytes:    34,
			sentBytes:        366,
		})

		require.Equal(t, upstreamALBHTTP, strings.TrimSpace(line))
	})
}
