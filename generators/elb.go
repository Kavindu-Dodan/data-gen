package generators

import (
	"fmt"
)

type ALBGen struct {
}

func (a *ALBGen) Get() ([]byte, error) {
	customizer := albCustomizer{
		logType:        randomALBType(),
		timestamp:      iso8601Now(),
		creationTime:   iso8601Now(),
		clientIPPort:   fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		targetIPPort:   fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		targetPortList: fmt.Sprintf("%s:%d", randomIP(), randomPort()),
		request:        `"GET http://www.example.com:80/ HTTP/1.1"`,
		elbID:          "appA/loadbalancer/123456789",
		targetARN:      "arn:aws:elasticloadbalancing:us-east-1:123456789:targetID",
		traceID:        `"trace=123456789"`,
		userAgent:      `"curl/7.46.0"`,
		requestID:      "TID_123456789",
	}

	return []byte(buildALBLogLine(customizer)), nil
}

func (a *ALBGen) ResetBatch() {
	// no-op
}

// helpers
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

const (
	requestProcessingTimeMs  = "0.000"
	targetProcessingTimeMs   = "0.001"
	responseProcessingTimeMs = "0.000"
	elbStatusCode            = "200"
	targetStatusCode         = "200"
	receivedBytes            = "34"
	sentBytes                = "366"
	matchedRulePriority      = "0"
	actionsExecuted          = `"forward"`
	targetStatusList         = `"200"`
	sslCipher                = "-"
	sslProtocol              = "-"
	domainName               = `"-"`
	chosenCertArn            = `"-"`
	redirectURL              = `"-"`
	errorReason              = `"-"`
	targetHealthReason       = `"-"`
	targetHealthDescription  = `"-"`
)

func buildALBLogLine(input albCustomizer) string {
	logLine := fmt.Sprintf(
		"%s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s\n",
		input.logType, input.timestamp, input.elbID,
		input.clientIPPort, input.targetIPPort,
		requestProcessingTimeMs, targetProcessingTimeMs, responseProcessingTimeMs,
		elbStatusCode, targetStatusCode,
		receivedBytes, sentBytes,
		input.request, input.userAgent,
		sslCipher, sslProtocol,
		input.targetARN, input.traceID,
		domainName, chosenCertArn,
		matchedRulePriority, input.creationTime,
		actionsExecuted, redirectURL, errorReason,
		input.targetPortList, targetStatusList,
		targetHealthReason, targetHealthDescription,
		input.requestID,
	)

	return logLine
}
