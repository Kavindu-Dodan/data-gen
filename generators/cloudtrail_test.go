package generators

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

// refer https://docs.aws.amazon.com/AmazonS3/latest/userguide/cloudtrail-logging-understanding-s3-entries.html#example-ct-log-s3
const sampleLog = `{
  "Records": [
    {
      "additionalEventData": {
        "AuthenticationMethod": "QueryString",
        "SignatureVersion": "SigV2",
        "aclRequired": "Yes"
      },
      "awsRegion": "us-west-2",
      "eventID": "cdc4b7ed-e171-4cef-975a-ad829d4123e8",
      "eventName": "ListBuckets",
      "eventSource": "s3.amazonaws.com",
      "eventTime": "2019-02-01T03:18:19Z",
      "eventType": "AwsApiCall",
      "eventVersion": "1.08",
      "recipientAccountId": "444455556666",
      "requestID": "47B8E8D397DCE7A6",
      "requestParameters": {
        "host": [
          "s3.us-west-2.amazonaws.com"
        ]
      },
      "sourceIPAddress": "127.0.0.1",
      "tlsDetails": {
        "cipherSuite": "ECDHE-RSA-AES128-GCM-SHA256",
        "clientProvidedHostHeader": "s3.amazonaws.com",
        "tlsVersion": "TLSv1.2"
      },
	  "resources": [
		{
			"accountId": "111122223333",
			"type": "AWS::S3::Bucket",
			"ARN": "arn:aws:s3:::abc"
		}
      ],
      "userAgent": "[]",
      "userIdentity": {
        "accessKeyId": "AKIAIOSFODNN7EXAMPLE",
        "accountId": "111122223333",
        "arn": "arn:aws:iam::111122223333:user/myUserName",
        "principalId": "111122223333",
        "type": "IAMUser",
        "userName": "myUserName"
      }
    }
  ]
}`

func Test_CloudTrailLog(t *testing.T) {
	customizer := cloudTrailCustomizer{
		AdditionalEventData: map[string]any{
			"SignatureVersion":     "SigV2",
			"AuthenticationMethod": "QueryString",
			"aclRequired":          "Yes",
		},
		awsRegion:          "us-west-2",
		eventID:            "cdc4b7ed-e171-4cef-975a-ad829d4123e8",
		eventName:          "ListBuckets",
		eventSource:        "s3.amazonaws.com",
		eventTime:          "2019-02-01T03:18:19Z",
		eventType:          "AwsApiCall",
		eventVersion:       "1.08",
		recipientAccountID: "444455556666",
		requestID:          "47B8E8D397DCE7A6",
		requestParameters: map[string]interface{}{
			"host": []string{"s3.us-west-2.amazonaws.com"},
		},
		resources: []any{
			map[string]any{
				"accountId": "111122223333",
				"type":      "AWS::S3::Bucket",
				"ARN":       "arn:aws:s3:::abc",
			},
		},
		sourceIPAddress: "127.0.0.1",
		tlsDetails: map[string]any{
			"tlsVersion":               "TLSv1.2",
			"cipherSuite":              "ECDHE-RSA-AES128-GCM-SHA256",
			"clientProvidedHostHeader": "s3.amazonaws.com",
		},
		userAgent: "[]",
		userIdentity: UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "111122223333",
			Arn:         "arn:aws:iam::111122223333:user/myUserName",
			AccountID:   "111122223333",
			AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
			UserName:    "myUserName",
		},
	}

	log := cloudTrailLogFor([]cloudTrailRecord{cloudTrailRecordFor(customizer)})
	generated, err := json.Marshal(log)
	require.NoError(t, err)

	var compareWith cloudTrailLog
	err = json.Unmarshal([]byte(sampleLog), &compareWith)
	require.NoError(t, err)

	marshal, err := json.Marshal(compareWith)
	require.NoError(t, err)

	require.Equal(t, marshal, generated)
}
