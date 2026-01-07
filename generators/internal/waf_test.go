package internal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// see - https://docs.aws.amazon.com/waf/latest/developerguide/logging-examples.html
const sample = `
{
   "timestamp":1683355579981,
   "webaclId": "111122223333:example-web-acl",
   "terminatingRuleId":"RateBasedRule",
   "terminatingRuleType":"RATE_BASED",
   "action":"BLOCK",
   "httpSourceName":"APIGW",
   "httpSourceId":"EXAMPLE11:rjvegx5guh:CanaryTest",
   "responseCodeSent":null,
   "httpRequest":{
      "clientIp":"52.46.82.45",
      "country":"FR",
      "headers":[
         {
            "name":"X-Forwarded-For",
            "value":"52.46.82.45"
         },
         {
            "name":"X-Forwarded-Proto",
            "value":"https"
         },
         {
            "name":"X-Forwarded-Port",
            "value":"443"
         },
         {
            "name":"Host",
            "value":"rjvegx5guh.execute-api.eu-west-3.amazonaws.com"
         }
      ],
      "uri":"/CanaryTest",
      "args":"",
      "httpVersion":"HTTP/1.1",
      "httpMethod":"GET",
      "requestId":"Ed0AiHF_CGYF-DA="
   }
}
`

func TestWafLogGenerating(t *testing.T) {
	customizer := wafCustomizer{
		timeStampMillis: 1683355579981,
		webACLID:        "111122223333:example-web-acl",
		ruleID:          "RateBasedRule",
		ruleType:        "RATE_BASED",
		action:          "BLOCK",
		httpSourceName:  "APIGW",
		httpSourceID:    "EXAMPLE11:rjvegx5guh:CanaryTest",
		httpRequest: wafHttpRequest{
			ClientIP: "52.46.82.45",
			Country:  "FR",
			Headers: []wafHttpHeader{
				{Name: "X-Forwarded-For", Value: "52.46.82.45"},
				{Name: "X-Forwarded-Proto", Value: "https"},
				{Name: "X-Forwarded-Port", Value: "443"},
				{Name: "Host", Value: "rjvegx5guh.execute-api.eu-west-3.amazonaws.com"},
			},
			URI:         "/CanaryTest",
			Args:        "",
			HTTPVersion: "HTTP/1.1",
			HTTPMethod:  "GET",
			RequestID:   "Ed0AiHF_CGYF-DA=",
		},
		responseCode: nil,
	}

	logLine := buildWAFLogLine(customizer)
	generated, err := json.Marshal(logLine)
	require.NoError(t, err)

	var compareWith wafLog
	err = json.Unmarshal([]byte(sample), &compareWith)
	require.NoError(t, err)

	marshal, err := json.Marshal(compareWith)
	require.NoError(t, err)

	require.Equal(t, marshal, generated)
}
