package generators

import (
	"fmt"
	"math/rand"
	"time"
)

type ALBGen struct {
}

func NewALBGen() *ALBGen {
	return &ALBGen{}
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
		"%s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s\n ",
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

// randomizers

func iso8601Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
}

func randomALBType() string {
	types := []string{"http", "https", "h2"}
	return types[rand.Intn(len(types))]
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	)
}

func randomPort() int {
	return rand.Intn(65535-1024) + 1024
}
