package generators

import (
	"fmt"
	"math/rand"
)

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
		logType:        randomSchema(),
		timestamp:      iso8601Now(),
		creationTime:   iso8601Now(),
		clientIPPort:   fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		targetIPPort:   fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		targetPortList: fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		request:        fmt.Sprintf("\"GET %s://%s:80/ HTTP/1.1\"", randomSchema(), randomDomain()),
		elbID:          fmt.Sprintf("appA/loadbalancer/%s", accountID),
		targetARN:      fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:targetID", randomRegion(), accountID),
		traceID:        fmt.Sprintf("\"trace=%d\"", rand.Intn(1000)),
		userAgent:      fmt.Sprintf("\"%s\"", userAgents[rand.Intn(len(userAgents))]),
		requestID:      randomAZ09String(5),
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

type albCustomizer struct {
	logType        string
	timestamp      string
	creationTime   string
	request        string
	elbID          string
	targetARN      string
	traceID        string
	userAgent      string
	clientIPPort   string
	targetIPPort   string
	targetPortList string
	requestID      string
}

func buildALBLogLine(input albCustomizer) string {
	logLine := fmt.Sprintf(
		"%s %s %s %s %s %f %f %f %s %s %d %d %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s\n",
		input.logType, input.timestamp, input.elbID,
		input.clientIPPort, input.targetIPPort,
		randomProcessingTime(), randomProcessingTime(), randomProcessingTime(),
		randomStatus(), randomStatus(),
		randomBytesSize(), randomBytesSize(),
		input.request, input.userAgent,
		randomSSLCipher(), sslProtocol(),
		input.targetARN, input.traceID,
		randomDomain(), chosenCertArn,
		matchedRulePriority, input.creationTime,
		actionsExecuted, redirectURL, errorReason,
		input.targetPortList, targetStatusList,
		targetHealthReason, targetHealthDescription,
		input.requestID,
	)

	return logLine
}
