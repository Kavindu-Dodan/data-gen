package generators

import (
	"fmt"
	"math/rand/v2"
)

const header = "version account-id interface-id srcaddr dstaddr srcport dstport protocol packets bytes start end action log-status"

type VPCGen struct {
	buf  trackedBuffer
	init bool
}

func newVPCGen() *VPCGen {
	return &VPCGen{
		buf:  newTrackedBuffer(),
		init: true,
	}
}

func (v *VPCGen) Generate() (int64, error) {
	var data []byte

	if v.init {
		v.init = false
		data = []byte(fmt.Sprintf("%s\n", header))
	}

	customizer := vpcCustomizer{
		Version:     2,
		AccountID:   randomAWSAccountID(),
		InterfaceID: "eni-123456789123",
		SrcAddr:     randomIP(),
		DstAddr:     randomIP(),
		SrcPort:     randomPort(),
		DstPort:     randomPort(),
		Protocol:    6, // TCP
		Packets:     rand.IntN(100) + 1,
		Bytes:       rand.IntN(1000) + 1,
		Start:       unixSeconds(),
		End:         unixSeconds(),
		Action:      randomVPCAction(),
		LogStatus:   "ok",
	}

	data = append(data, []byte(buildVPCLogLine(customizer))...)
	err := v.buf.write(data)
	return v.buf.size(), err
}

func (v *VPCGen) GetAndReset() []byte {
	v.init = true
	return v.buf.getAndRest()
}

type vpcCustomizer struct {
	Version     int
	AccountID   string
	InterfaceID string
	SrcAddr     string
	DstAddr     string
	SrcPort     int
	DstPort     int
	Protocol    int
	Packets     int
	Bytes       int
	Start       int64
	End         int64
	Action      string
	LogStatus   string
}

func buildVPCLogLine(vpcCustomizer vpcCustomizer) string {
	return fmt.Sprintf(
		"%d %s %s %s %s %d %d %d %d %d %d %d %s %s\n",
		vpcCustomizer.Version,
		vpcCustomizer.AccountID,
		vpcCustomizer.InterfaceID,
		vpcCustomizer.SrcAddr,
		vpcCustomizer.DstAddr,
		vpcCustomizer.SrcPort,
		vpcCustomizer.DstPort,
		vpcCustomizer.Protocol,
		vpcCustomizer.Packets,
		vpcCustomizer.Bytes,
		vpcCustomizer.Start,
		vpcCustomizer.End,
		vpcCustomizer.Action,
		vpcCustomizer.LogStatus,
	)
}
