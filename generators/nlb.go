package generators

import (
	"fmt"
	"math/rand"
)

type NLBgen struct {
	buf trackedBuffer
}

func NewNLBGen() *NLBgen {
	return &NLBgen{
		buf: newTrackedBuffer(),
	}
}

func (a *NLBgen) Generate() (int64, error) {
	customizer := nlbCustomizer{
		time:              iso8601Now(),
		name:              fmt.Sprintf("nlb/my-nlb/%s", randomID()),
		elbID:             randomID(),
		clientIPPort:      fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		destinationIPPort: fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		conMs:             rand.Intn(1000),
		tlsHSMs:           rand.Intn(1000),
		receivedBytes:     randomBytesSize(),
		sentBytes:         randomBytesSize(),
		tlsAlert:          "-",
		certARN:           randomCertArn(),
		cipher:            randomSSLCipher(),
		protocol:          sslProtocol(),
		domain:            randomDomain(),
		feProtocol:        "-",
		beProtocol:        "-",
		alpnList:          "-",
		creationTime:      iso8601Now(),
	}

	err := a.buf.write([]byte(buildNLBLogLine(customizer)))
	if err != nil {
		return 0, err
	}

	return a.buf.size(), err
}

func (a *NLBgen) GetAndReset() []byte {
	return a.buf.getAndReset()
}

// helpers

type nlbCustomizer struct {
	time              string
	name              string
	elbID             string
	clientIPPort      string
	destinationIPPort string
	conMs             int
	tlsHSMs           int
	receivedBytes     int
	sentBytes         int
	tlsAlert          string
	certARN           string
	cipher            string
	protocol          string
	domain            string
	feProtocol        string
	beProtocol        string
	alpnList          string
	creationTime      string
}

func buildNLBLogLine(input nlbCustomizer) string {
	// Construct the log line
	log := fmt.Sprintf(
		"%s %s %s %s %s %s %s %d %d %d %d %s %s %s %s %s %s %s %s %s %s %s\n",
		"tls", "2.0", input.time, input.name, input.elbID, input.clientIPPort, input.destinationIPPort, input.conMs,
		input.tlsHSMs, input.receivedBytes, input.sentBytes,
		input.tlsAlert, input.certARN, "-", // Placeholder for future fields
		input.cipher, input.protocol, "-", // Placeholder for future fields
		input.domain,
		input.feProtocol, input.beProtocol, input.alpnList, input.creationTime,
	)

	return log
}
