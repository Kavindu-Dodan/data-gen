package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// refer example of https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs-records-examples.html#flow-log-example-accepted-rejected
const acceptedVPC = "2 123456789010 eni-1235b8ca123456789 172.31.16.139 172.31.16.21 20641 22 6 20 4249 1418530010 1418530070 ACCEPT OK"
const rejectedVPC = "2 123456789010 eni-1235b8ca123456789 172.31.9.69 172.31.9.12 49761 3389 6 20 4249 1418530010 1418530070 REJECT OK"

func Test_buildVPC(t *testing.T) {
	t.Run("Validate AWS documented VPC - Accepted", func(t *testing.T) {
		line := buildVPCLogLine(vpcCustomizer{
			Version:     2,
			AccountID:   "123456789010",
			InterfaceID: "eni-1235b8ca123456789",
			SrcAddr:     "172.31.16.139",
			DstAddr:     "172.31.16.21",
			SrcPort:     20641,
			DstPort:     22,
			Protocol:    6,
			Packets:     20,
			Bytes:       4249,
			Start:       1418530010,
			End:         1418530070,
			Action:      "ACCEPT",
			LogStatus:   "OK",
		})

		require.Equal(t, acceptedVPC, strings.TrimSpace(line))
	})

	t.Run("Validate AWS documented VPC - Accepted", func(t *testing.T) {
		line := buildVPCLogLine(vpcCustomizer{
			Version:     2,
			AccountID:   "123456789010",
			InterfaceID: "eni-1235b8ca123456789",
			SrcAddr:     "172.31.9.69",
			DstAddr:     "172.31.9.12",
			SrcPort:     49761,
			DstPort:     3389,
			Protocol:    6,
			Packets:     20,
			Bytes:       4249,
			Start:       1418530010,
			End:         1418530070,
			Action:      "REJECT",
			LogStatus:   "OK",
		})

		require.Equal(t, rejectedVPC, strings.TrimSpace(line))
	})
}
