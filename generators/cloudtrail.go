package generators

import (
	"encoding/json"
	"github.com/google/uuid"
	"math/rand/v2"
)

type cloudTrail struct {
	current     []cloudTrailRecord
	currentSize int64
}

func newCloudTrailGen() *cloudTrail {
	return &cloudTrail{
		current: []cloudTrailRecord{},
	}
}

func (c *cloudTrail) Generate() (int64, error) {
	id := randomAZ09String(12)
	accountID := randomSampleAccountID()
	s3EventName := randomS3EventName()
	parameters, resources := generateRequestAndResource(s3EventName, accountID)
	responseElements := map[string]any{
		"requestId": id,
		"kmsKeyId":  "arn:aws:kms:us-east-1:123456789012:key/" + randomAZ09String(1),
	}

	customizer := cloudTrailCustomizer{
		awsRegion:          randomRegion(),
		eventCategory:      eventCategory,
		eventID:            uuid.NewString(),
		eventName:          s3EventName,
		eventSource:        s3EventSource,
		eventTime:          iso8601Now(),
		eventType:          eventType,
		eventVersion:       eventVersion,
		managementEvent:    &ff,
		readOnly:           &ff,
		recipientAccountID: accountID,
		requestID:          id,
		requestParameters:  parameters,
		resources:          []any{resources},
		responseElements:   responseElements,
		sharedEventID:      randomAZ09String(16),
		sourceIPAddress:    randomIP(),
		userAgent:          randomUserAgent(),
		userIdentity:       ctUserIdentity(),
		tlsDetails: map[string]any{
			"tlsVersion":  randomTLSProtocol(),
			"cipherSuite": randomSSLCipher(),
		},
	}

	// 10% chance of error
	if rand.IntN(10) < 1 {
		customizer.errorCode, customizer.errorMessage = randomErrorCodeAndMessage()
	}

	newRecord := cloudTrailRecordFor(customizer)
	c.current = append(c.current, newRecord)

	size, err := json.Marshal(newRecord)
	if err != nil {
		return 0, err
	}

	c.currentSize += int64(len(size))
	return c.currentSize, nil
}

func (c *cloudTrail) GetAndReset() []byte {
	marshal, _ := json.Marshal(cloudTrailLogFor(c.current))
	c.current = []cloudTrailRecord{} // Reset current records
	c.currentSize = 0                // Reset current size
	return marshal
}

// helpers

const (
	eventVersion  = "1.11"
	s3EventSource = "s3.amazonaws.com"
	eventType     = "AwsApiCall"
	eventCategory = "Data"
)

type cloudTrailCustomizer struct {
	AdditionalEventData map[string]any
	awsRegion           string
	errorCode           string
	errorMessage        string
	eventCategory       string
	eventID             string
	eventName           string
	eventSource         string
	eventTime           string
	eventType           string
	eventVersion        string
	insightDetails      map[string]any
	managementEvent     *bool
	readOnly            *bool
	recipientAccountID  string
	requestID           string
	requestParameters   map[string]any
	resources           []any
	responseElements    map[string]any
	sharedEventID       string
	sourceIPAddress     string
	tlsDetails          map[string]any
	userAgent           string
	userIdentity        UserIdentity
}

type cloudTrailRecord struct {
	AdditionalEventData          map[string]interface{} `json:"additionalEventData,omitempty"`
	AwsRegion                    string                 `json:"awsRegion,omitempty"`
	ErrorCode                    string                 `json:"errorCode,omitempty"`
	ErrorMessage                 string                 `json:"errorMessage,omitempty"`
	EventCategory                string                 `json:"eventCategory,omitempty"`
	EventID                      string                 `json:"eventID,omitempty"`
	EventName                    string                 `json:"eventName,omitempty"`
	EventSource                  string                 `json:"eventSource,omitempty"`
	EventTime                    string                 `json:"eventTime,omitempty"`
	EventType                    string                 `json:"eventType,omitempty"`
	EventVersion                 string                 `json:"eventVersion,omitempty"`
	InsightDetails               map[string]any         `json:"insightDetails,omitempty"`
	ManagementEvent              *bool                  `json:"managementEvent,omitempty"`
	ReadOnly                     *bool                  `json:"readOnly,omitempty"`
	RecipientAccountID           string                 `json:"recipientAccountId,omitempty"`
	RequestID                    string                 `json:"requestID,omitempty"`
	RequestParameters            map[string]any         `json:"requestParameters,omitempty"`
	Resources                    []any                  `json:"resources,omitempty"`
	ResponseElements             map[string]any         `json:"responseElements,omitempty"`
	SessionCredentialFromConsole string                 `json:"sessionCredentialFromConsole,omitempty"`
	SharedEventID                string                 `json:"sharedEventID,omitempty"`
	SourceIPAddress              string                 `json:"sourceIPAddress,omitempty"`
	TLSDetails                   map[string]any         `json:"tlsDetails"`
	UserAgent                    string                 `json:"userAgent,omitempty"`
	UserIdentity                 UserIdentity           `json:"userIdentity"`
}

type UserIdentity struct {
	Type        string `json:"type,omitempty"`
	PrincipalID string `json:"principalId,omitempty"`
	Arn         string `json:"arn,omitempty"`
	AccountID   string `json:"accountId,omitempty"`
	AccessKeyID string `json:"accessKeyId,omitempty"`
	UserName    string `json:"userName,omitempty"`
	InvokedBy   string `json:"invokedBy,omitempty"`
}

type cloudTrailLog struct {
	Records []cloudTrailRecord `json:"Records"`
}

func cloudTrailRecordFor(customizer cloudTrailCustomizer) cloudTrailRecord {
	return cloudTrailRecord{
		AdditionalEventData: customizer.AdditionalEventData,
		AwsRegion:           customizer.awsRegion,
		ErrorCode:           customizer.errorCode,
		ErrorMessage:        customizer.errorMessage,
		EventCategory:       customizer.eventCategory,
		EventID:             customizer.eventID,
		EventName:           customizer.eventName,
		EventSource:         customizer.eventSource,
		EventTime:           customizer.eventTime,
		EventType:           customizer.eventType,
		EventVersion:        customizer.eventVersion,
		InsightDetails:      customizer.insightDetails,
		ManagementEvent:     customizer.managementEvent,
		ReadOnly:            customizer.readOnly,
		RecipientAccountID:  customizer.recipientAccountID,
		RequestID:           customizer.requestID,
		RequestParameters:   customizer.requestParameters,
		Resources:           customizer.resources,
		ResponseElements:    customizer.responseElements,
		SharedEventID:       customizer.sharedEventID,
		SourceIPAddress:     customizer.sourceIPAddress,
		TLSDetails:          customizer.tlsDetails,
		UserAgent:           customizer.userAgent,
		UserIdentity:        customizer.userIdentity,
	}
}

func cloudTrailLogFor(records []cloudTrailRecord) cloudTrailLog {
	return cloudTrailLog{
		Records: records,
	}
}
