package generators

import (
	"encoding/json"
	"time"
)

// WAFGen generates AWS WAF logs in JSON format.
type WAFGen struct {
	wafId string
	buf   trackedBuffer
}

func newWAFGen() *WAFGen {
	return &WAFGen{
		wafId: randomWAFACLID(),
	}
}

func (w *WAFGen) Generate() (int64, error) {
	customizer := wafCustomizer{
		timeStampMillis: time.Now().UnixMilli(),
		webACLID:        w.wafId,
		ruleID:          randomWAFRuleId(),
		ruleType:        randomWafRuleType(),
		action:          randomWafAction(),
		httpSourceName:  randomWafSourceName(),
		httpSourceID:    randomSourceID(), // Note - This does not match with actual ID format
		httpRequest: wafHttpRequest{
			ClientIP:    randomIP(),
			Country:     randomCountryCode(),
			Headers:     randomWafHeaders(),
			URI:         randomURIPath(),
			Args:        randomQueryString(),
			HTTPVersion: "HTTP/1.1",
			HTTPMethod:  randomHTTPMethod(),
			RequestID:   randomAZ09String(8),
			Fragment:    randomFragment(),
			Scheme:      randomSchema(),
			Host:        "example.com",
		},
	}

	logLine := buildWAFLogLine(customizer)
	marshaled, err := json.Marshal(logLine)
	if err != nil {
		return 0, err
	}

	// WAF logs are newline delimited
	marshaled = append(marshaled, '\n')

	err = w.buf.write(marshaled)
	if err != nil {
		return 0, err
	}

	return w.buf.size(), nil
}

func (w *WAFGen) GetAndReset() []byte {
	return w.buf.getAndReset()
}

// wafCustomizer holds parameters for generating a WAF log entry.
type wafCustomizer struct {
	timeStampMillis int64
	webACLID        string
	ruleID          string
	ruleType        string
	action          string
	httpSourceName  string
	httpSourceID    string
	httpRequest     wafHttpRequest
	responseCode    *int
}

// wafHttpRequest contains HTTP request details captured by WAF.
type wafHttpRequest struct {
	ClientIP    string          `json:"clientIp"`
	Country     string          `json:"country"`
	Headers     []wafHttpHeader `json:"headers"`
	URI         string          `json:"uri"`
	Args        string          `json:"args"`
	HTTPVersion string          `json:"httpVersion"`
	HTTPMethod  string          `json:"httpMethod"`
	RequestID   string          `json:"requestID"`
	Fragment    string          `json:"fragment"`
	Scheme      string          `json:"scheme"`
	Host        string          `json:"host"`
}

// wafHttpHeader represents an HTTP header name-value pair.
type wafHttpHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// wafLog represents the complete WAF log structure.
// Credits: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/encoding/awslogsencodingextension/internal/unmarshaler/waf/unmarshaler.go
type wafLog struct {
	Timestamp           int64          `json:"timestamp"`
	WebACLID            string         `json:"webaclId"`
	TerminatingRuleID   string         `json:"terminatingRuleId"`
	TerminatingRuleType string         `json:"terminatingRuleType"`
	Action              string         `json:"action"`
	HTTPSourceName      string         `json:"httpSourceName"`
	HTTPSourceID        string         `json:"httpSourceId"`
	ResponseCodeSent    *int           `json:"responseCodeSent"`
	HTTPRequest         wafHttpRequest `json:"httpRequest"`
}

func buildWAFLogLine(c wafCustomizer) wafLog {
	return wafLog{
		Timestamp:           c.timeStampMillis,
		WebACLID:            c.webACLID,
		TerminatingRuleID:   c.ruleID,
		TerminatingRuleType: c.ruleType,
		Action:              c.action,
		HTTPSourceName:      c.httpSourceName,
		HTTPSourceID:        c.httpSourceID,
		HTTPRequest:         c.httpRequest,
		ResponseCodeSent:    c.responseCode,
	}
}
