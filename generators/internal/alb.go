package internal

import (
	"fmt"
	"math/rand"
)

// ALBGen generates AWS Application Load Balancer access logs in standard format.
type ALBGen struct {
	buf trackedBuffer
}

func NewALBGen() *ALBGen {
	return &ALBGen{
		buf: newTrackedBuffer(),
	}
}

func (a *ALBGen) Generate() (int64, error) {
	accountID := randomSampleAccountID()

	customizer := albCustomizer{
		cipher:           randomSSLCipher(),
		clientIPPort:     fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		creationTime:     iso8601Now(),
		domain:           randomDomain(),
		elbID:            fmt.Sprintf("appA/loadbalancer/%s", accountID),
		elbStatus:        randomStatus(),
		logType:          randomSchema(),
		receivedBytes:    randomBytesSize(),
		request:          fmt.Sprintf("\"GET %s://%s:80/ HTTP/1.1\"", randomSchema(), randomDomain()),
		requestID:        randomAZ09String(5),
		requestProcTime:  randomProcessingTime(),
		responseProcTime: randomProcessingTime(),
		sentBytes:        randomBytesSize(),
		sslProtocol:      randomTLSProtocol(),
		targetARN:        fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:targetID", randomRegion(), accountID),
		targetIPPort:     fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		targetPortList:   fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		targetProcTime:   randomProcessingTime(),
		targetStatus:     randomStatus(),
		timestamp:        iso8601Now(),
		traceID:          fmt.Sprintf("\"trace=%d\"", rand.Intn(1000)),
		userAgent:        fmt.Sprintf("\"%s\"", userAgents[rand.Intn(len(userAgents))]),
	}

	err := a.buf.write([]byte(buildALBLogLine(customizer)))
	if err != nil {
		return 0, err
	}

	return a.buf.size(), err
}

func (a *ALBGen) GetAndReset() []byte {
	return a.buf.getAndReset()
}

// helpers

const (
	matchedRulePriority     = "0"
	actionsExecuted         = `"forward"`
	targetStatusList        = `"200"`
	chosenCertArn           = `"-"`
	redirectURL             = `"-"`
	errorReason             = `"-"`
	targetHealthReason      = `"-"`
	targetHealthDescription = `"-"`
)

// albCustomizer holds all fields needed to construct an ALB log entry.
type albCustomizer struct {
	cipher           string
	clientIPPort     string
	creationTime     string
	domain           string
	elbID            string
	elbStatus        string
	logType          string
	receivedBytes    int
	request          string
	requestID        string
	requestProcTime  float32
	responseProcTime float32
	sentBytes        int
	sslProtocol      string
	targetARN        string
	targetIPPort     string
	targetPortList   string
	targetProcTime   float32
	targetStatus     string
	timestamp        string
	traceID          string
	userAgent        string
}

func buildALBLogLine(input albCustomizer) string {
	logLine := fmt.Sprintf(
		"%s %s %s %s %s %0.3f %0.3f %0.3f %s %s %d %d %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s\n",
		input.logType, input.timestamp, input.elbID,
		input.clientIPPort, input.targetIPPort,
		input.requestProcTime, input.targetProcTime, input.responseProcTime,
		input.elbStatus, input.targetStatus,
		input.receivedBytes, input.sentBytes,
		input.request, input.userAgent,
		input.cipher, input.sslProtocol,
		input.targetARN, input.traceID,
		input.domain, chosenCertArn,
		matchedRulePriority, input.creationTime,
		actionsExecuted, redirectURL, errorReason,
		input.targetPortList, targetStatusList,
		targetHealthReason, targetHealthDescription,
		input.requestID,
	)

	return logLine
}
