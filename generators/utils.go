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
var albTypes = []string{"http", "https"}
var vpcActions = []string{"ACCEPT", "REJECT"}
var sampleAccountIDs = []string{"123456789012", "987654321098", "111122223333", "444455556666", "777788889999"}
var samplePrincipalIDs = []string{"AID1234567890", "AID0987654321", "AID1111222233", "AID7777888899"}
var regions = []string{"us-east-1", "us-west-1", "us-west-2", "eu-west-1", "eu-central-1"}
var s3EventNames = []string{"PutObject", "GetObject", "DeleteObject", "ListObjects"}

func iso8601Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
}

func unixSeconds() int64 {
	return time.Now().Unix()
}

func randomALBType() string {
	return albTypes[rand.Intn(len(albTypes))]
}

func randomVPCAction() string {
	return vpcActions[rand.Intn(len(vpcActions))]
}

// randomIP from 72.16.101.0/24
func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		172,
		16,
		101,
		rand.Intn(256),
	)
}

// randomPort in range 58080 to 62000
func randomPort() int {
	return rand.Intn(62000-58080) + 58080
}

func randomAWSAccountID() string {
	// Generate a 12-digit number, ensuring the first digit is not 0
	accountID := rand.Intn(9) + 1
	for i := 1; i < 12; i++ {
		accountID = accountID*10 + rand.Intn(10)
	}
	return fmt.Sprintf("%012d", accountID)
}

func ctUserIdentity() UserIdentity {
	userName := fmt.Sprintf("user%d", rand.Intn(1000))
	accountID := randomSampleAccountID()

	arn := randomIAMArn(accountID, userName)

	return UserIdentity{
		Type:        "IAMUser",
		PrincipalID: samplePrincipalIDs[rand.Intn(len(samplePrincipalIDs))],
		Arn:         arn,
		AccountID:   accountID,
		AccessKeyID: randomAZaz09String(20),
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
	return randomAZaz09String(4) + "-" + randomAZaz09String(4)
}

func randomS3ObjectKey() string {
	return "object_" + randomAZaz09String(5) + ".txt"
}
