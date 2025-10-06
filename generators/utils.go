package generators

import (
	"fmt"
	"math/rand"
	"time"
)

// randomizers

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const charsCapital = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var ff = false
var tt = true

var contentTypes = []string{"text/html", "application/json", "text/plain", "application/xml"}
var countryCodes = []string{"US", "GB", "DE", "FR", "IN", "CN", "JP", "AU", "CA", "BR"}
var sampleDomains = []string{"example.com", "test.com", "sample.org", "demo.net"}
var httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
var httpSchema = []string{"http", "https"}
var httpSourceIDs = []string{"E2A1BCD34FGH56", "E3B2CDE45GHI67", "E4C3DEF56HIJ78", "E5D4EFG67IJK89"}
var queryStrings = []string{"", "a=1&b=2", "user=abc", "id=12345", "page=1", "search=term"}
var randomFragments = []string{"", "#browser", "#app"}
var randomPhrases = []string{"some random phrase", "another random phrase", "yet another random phrase", "log on something"}
var regions = []string{"us-east-1", "us-west-1", "us-west-2", "eu-west-1", "eu-central-1"}
var s3EventNames = []string{"PutObject", "GetObject", "DeleteObject", "ListObjects"}
var sampleAccountIDs = []string{"123456789012", "987654321098", "111122223333", "444455556666", "777788889999"}
var samplePrincipalIDs = []string{"AID1234567890", "AID0987654321", "AID1111222233", "AID7777888899"}
var sampleRuleIDs = []string{"rule-1", "rule-2", "rule-3", "rule-4", "rule-5"}
var sslCiphers = []string{"ECDHE-RSA-AES128-GCM-SHA256", "ECDHE-RSA-AES256-GCM-SHA384", "AES128-GCM-SHA256"}
var sslProtocols = []string{"TLSv1.2", "TLSv1.3"}
var statuses = []string{"200", "400", "500"}
var uriPaths = []string{"/", "/home", "/api/resource", "/login"}
var userAgents = []string{"Mozilla/5.0, AppleWebKit/537.36, Chrome/58.0.3029.110, Safari/537.3", "curl/7.46.0"}
var uuids = []string{"550e8400-e29b-41d4-a716-446655440000", "123e4567-e89b-12d3-a456-426614174000", "9b2c3d4e-5f6a-7b8c-9d0e-1f2a3b4c5d6e"}
var vpcActions = []string{"ACCEPT", "REJECT"}
var wafActions = []string{"ALLOW", "BLOCK", "COUNT"}
var wafRuleTypes = []string{"REGULAR", "RATE_BASED", "GROUP"}
var wafSampleHTTPSourceNames = []string{"ALB", "CloudFront", "API Gateway"}

// ipPrefix contains example public IP address prefixes, representing geo-distributed or commonly used public IP ranges.
var ipPrefix = []int{1, 8, 31, 41, 91, 123, 179, 201, 210, 250}

func randomDomain() string {
	return sampleDomains[rand.Intn(len(sampleDomains))]
}

func iso8601Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
}

func unixSeconds(delay int) int64 {
	return time.Now().Unix() + int64(delay)
}

// randomProcessingTime returns a random float as string between 0.500 and 1.499
func randomProcessingTime() float32 {
	return (0.5 + rand.Float32()*1000) / 1000
}

func randomStatus() string {
	return statuses[rand.Intn(len(statuses))]
}

func randomBytesSize() int {
	return rand.Intn(5000-200) + 200
}

func randomSchema() string {
	return httpSchema[rand.Intn(len(httpSchema))]
}

func sslProtocol() string {
	return sslProtocols[rand.Intn(len(sslProtocols))]
}

func randomSSLCipher() string {
	return sslCiphers[rand.Intn(len(sslCiphers))]
}

func randomVPCAction() string {
	return vpcActions[rand.Intn(len(vpcActions))]
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ipPrefix[rand.Intn(len(ipPrefix))],
		1,
		1,
		1,
	)
}

// randomPort in range 58080 to 59090
func randomPort() int {
	return rand.Intn(59090-58080) + 58080
}

func ctUserIdentity() UserIdentity {
	userName := fmt.Sprintf("user%d", rand.Intn(10))
	accountID := randomSampleAccountID()

	arn := randomIAMArn(accountID, userName)

	return UserIdentity{
		Type:        "IAMUser",
		PrincipalID: samplePrincipalIDs[rand.Intn(len(samplePrincipalIDs))],
		Arn:         arn,
		AccountID:   accountID,
		AccessKeyID: "AKIA" + randomAZ09String(4),
		UserName:    userName,
	}
}

func randomIAMArn(accId string, user string) string {
	return fmt.Sprintf("arn:aws:iam::%s:user/%s", accId, user)
}

func randomAZaz09String(size int) string {
	key := make([]byte, size)
	for i := range key {
		key[i] = charset[rand.Intn(len(charset))]
	}
	return string(key)
}

func randomAZ09String(size int) string {
	key := make([]byte, size)
	for i := range key {
		key[i] = charsCapital[rand.Intn(len(charsCapital))]
	}
	return string(key)
}

func randomSampleAccountID() string {
	return sampleAccountIDs[rand.Intn(len(sampleAccountIDs))]
}

func randomRegion() string {
	return regions[rand.Intn(len(regions))]
}

func randomS3EventName() string {
	return s3EventNames[rand.Intn(len(s3EventNames))]
}

func generateRequestAndResource(eventName string, accID string) (map[string]any, map[string]any) {
	bucket := randomBucketName()
	s3ObjectKey := randomS3ObjectKey()

	switch eventName {
	case "PutObject", "GetObject", "DeleteObject":

		request := map[string]any{
			"bucketName": bucket,
			"key":        s3ObjectKey,
		}

		resource := map[string]any{
			"ARN":       fmt.Sprintf("arn:aws:s3:::%s/%s", bucket, s3ObjectKey),
			"type":      "AWS::S3::Object",
			"accountId": accID,
		}

		return request, resource
	case "ListObjects":
		request := map[string]any{
			"bucketName": bucket,
			"maxKeys":    1000,
		}

		resource := map[string]any{
			"ARN":       fmt.Sprintf("arn:aws:s3:::%s", bucket),
			"type":      "AWS::S3::Bucket",
			"accountId": accID,
		}

		return request, resource
	default:
		request := map[string]any{
			"raw": map[string]any{
				"eventName": eventName,
			},
		}

		resource := map[string]any{
			"ARN": fmt.Sprintf("arn:aws:s3:::%s", bucket),
		}

		return request, resource
	}
}

func randomBucketName() string {
	return "bucket-" + fmt.Sprintf("%03d", rand.Intn(1000))
}

func randomS3ObjectKey() string {
	return "object_" + randomAZaz09String(2) + ".txt"
}

func randomFragment() string {
	return randomFragments[rand.Intn(len(randomFragments))]
}

func randomHTTPMethod() string {
	return httpMethods[rand.Intn(len(httpMethods))]
}

func randomSourceID() string {
	return httpSourceIDs[rand.Intn(len(httpSourceIDs))]
}

func randomQueryString() string {
	return queryStrings[rand.Intn(len(queryStrings))]
}

func randomURIPath() string {
	return uriPaths[rand.Intn(len(uriPaths))]
}

func randomWafHeaders() []wafHttpHeader {
	cType := wafHttpHeader{
		Name:  "Content-Type",
		Value: contentTypes[rand.Intn(len(contentTypes))],
	}

	uAgent := wafHttpHeader{
		Name:  "User-Agent",
		Value: userAgents[rand.Intn(len(userAgents))],
	}

	accept := wafHttpHeader{
		Name:  "Accept",
		Value: "*/*",
	}

	keepAlive := wafHttpHeader{
		Name:  "Connection",
		Value: "keep-alive",
	}

	return []wafHttpHeader{cType, uAgent, accept, keepAlive}
}

func randomWAFRuleId() string {
	return sampleRuleIDs[rand.Intn(len(sampleRuleIDs))]
}

func randomCountryCode() string {
	return countryCodes[rand.Intn(len(countryCodes))]
}

func randomWafRuleType() string {
	return wafRuleTypes[rand.Intn(len(wafRuleTypes))]
}

func randomWafAction() string {
	return wafActions[rand.Intn(len(wafActions))]
}

func randomWAFACLID() string {
	return fmt.Sprintf("arn:aws:wafv2:%s:%s:regional/webacl/sample-web-acl/%s", randomRegion(), randomSampleAccountID(), uuids[rand.Intn(len(uuids))])
}

func randomWafSourceName() string {
	return wafSampleHTTPSourceNames[rand.Intn(len(wafSampleHTTPSourceNames))]
}

func randomLogString(size int) string {
	var buildBytes []byte
	for len(buildBytes) < size {
		buildBytes = append(buildBytes, []byte(randomPhrases[rand.Intn(len(randomPhrases))])...)
		buildBytes = append(buildBytes, ' ')
	}

	return string(buildBytes)
}
