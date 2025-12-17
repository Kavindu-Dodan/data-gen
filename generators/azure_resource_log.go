package generators

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"data-gen/conf"
)

// AzureResourceLogGen generates Azure resource logs in JSON format.
type AzureResourceLogGen struct {
	cfg azureResourceLogCfg
	buf trackedBuffer
}

// azureResourceLogCfg specifies the Azure resource log generator configuration.
type azureResourceLogCfg struct {
	RecordsPerBatch int `yaml:"records_per_batch"`
}

func newDefaultAzureResourceLogCfg() *azureResourceLogCfg {
	return &azureResourceLogCfg{
		RecordsPerBatch: 1,
	}
}

func newAzureResourceLogGen(cfg conf.InputConfig) *AzureResourceLogGen {
	config := newDefaultAzureResourceLogCfg()
	err := cfg.Conf.Decode(&config)
	if err != nil {
		// If no config is provided or error, use defaults
		config = newDefaultAzureResourceLogCfg()
	}

	// load env variable overrides if any
	if v := os.Getenv("ENV_AZURE_RECORDS_PER_BATCH"); v != "" {
		if count, err := strconv.Atoi(v); err == nil && count > 0 {
			config.RecordsPerBatch = count
		}
	}

	return &AzureResourceLogGen{
		cfg: *config,
		buf: newTrackedBuffer(),
	}
}

func (a *AzureResourceLogGen) Generate() (int64, error) {
	// Generate multiple log entries based on configuration
	records := make([]azureResourceLog, a.cfg.RecordsPerBatch)
	for i := 0; i < a.cfg.RecordsPerBatch; i++ {
		records[i] = buildAzureResourceLog()
	}

	// Wrap the log entries in a records field (Azure standard format)
	wrapper := map[string]interface{}{
		"records": records,
	}

	marshaled, err := json.Marshal(wrapper)
	if err != nil {
		return 0, err
	}

	// Azure resource logs are newline delimited
	marshaled = append(marshaled, '\n')

	err = a.buf.write(marshaled)
	if err != nil {
		return 0, err
	}

	return a.buf.size(), nil
}

func (a *AzureResourceLogGen) GetAndReset() []byte {
	return a.buf.getAndReset()
}

// azureResourceLog represents an Azure resource log entry.
// Based on common Azure resource log schema
type azureResourceLog struct {
	Time              string                 `json:"time"`
	ResourceID        string                 `json:"resourceId"`
	OperationName     string                 `json:"operationName"`
	Category          string                 `json:"category"`
	ResultType        string                 `json:"resultType,omitempty"`
	ResultSignature   string                 `json:"resultSignature,omitempty"`
	ResultDescription string                 `json:"resultDescription,omitempty"`
	DurationMs        int                    `json:"durationMs,omitempty"`
	CallerIPAddress   string                 `json:"callerIpAddress,omitempty"`
	CorrelationID     string                 `json:"correlationId,omitempty"`
	Identity          *azureIdentity         `json:"identity,omitempty"`
	Level             string                 `json:"Level"`
	Location          string                 `json:"location"`
	Properties        map[string]interface{} `json:"properties,omitempty"`
}

// azureIdentity represents the identity associated with the Azure operation.
type azureIdentity struct {
	Authorization *azureAuthorization `json:"authorization,omitempty"`
	Claims        map[string]string   `json:"claims,omitempty"`
}

// azureAuthorization contains authorization details for the Azure operation.
type azureAuthorization struct {
	Scope  string `json:"scope,omitempty"`
	Action string `json:"action,omitempty"`
	Role   string `json:"role,omitempty"`
}

func buildAzureResourceLog() azureResourceLog {
	category := randomAzureCategory()
	operationName := randomAzureOperationName(category)
	resultType := randomAzureResultType()

	log := azureResourceLog{
		Time:              time.Now().UTC().Format(time.RFC3339Nano),
		ResourceID:        randomAzureResourceID(),
		OperationName:     operationName,
		Category:          category,
		ResultType:        resultType,
		DurationMs:        randomDurationMs(),
		CallerIPAddress:   randomIP(),
		CorrelationID:     randomAzureGUID(),
		Level:             randomAzureLogLevel(),
		Location:          randomAzureRegion(),
	}

	// Add result signature for failed operations
	if resultType != "Success" {
		log.ResultSignature = randomAzureErrorCode()
		log.ResultDescription = randomAzureErrorDescription()
	}

	// Add identity for certain operation types
	if shouldHaveIdentity(operationName) {
		log.Identity = &azureIdentity{
			Authorization: &azureAuthorization{
				Scope:  randomAzureResourceID(),
				Action: randomAzureAction(),
				Role:   randomAzureRole(),
			},
			Claims: map[string]string{
				"aud": "https://management.azure.com/",
				"iss": "https://sts.windows.net/" + randomAzureGUID() + "/",
				"iat": "1234567890",
				"nbf": "1234567890",
				"exp": "1234567890",
			},
		}
	}

	// Add category-specific properties
	log.Properties = generateAzureProperties(category, operationName)

	return log
}

func shouldHaveIdentity(operationName string) bool {
	// Add identity for write operations
	return operationName == "Microsoft.Storage/storageAccounts/write" ||
		operationName == "Microsoft.Compute/virtualMachines/write" ||
		operationName == "Microsoft.Network/networkSecurityGroups/securityRules/write" ||
		operationName == "Microsoft.KeyVault/vaults/secrets/write"
}

func generateAzureProperties(category, operationName string) map[string]interface{} {
	props := make(map[string]interface{})

	switch category {
	case "Administrative":
		props["eventCategory"] = "Administrative"
		props["eventDataId"] = randomAzureGUID()
		props["operationId"] = randomAzureGUID()
		props["httpRequest"] = map[string]interface{}{
			"clientRequestId": randomAzureGUID(),
			"clientIpAddress": randomIP(),
			"method":          randomHTTPMethod(),
		}
	case "Security":
		props["securityEventType"] = randomAzureSecurityEventType()
		props["protocol"] = randomAzureProtocol()
		props["direction"] = randomAzureDirection()
	case "ServiceHealth":
		props["eventType"] = "ServiceIssue"
		props["trackingId"] = randomAzureGUID()
		props["incidentType"] = randomAzureIncidentType()
	case "ResourceHealth":
		props["currentHealthStatus"] = randomAzureHealthStatus()
		props["previousHealthStatus"] = randomAzureHealthStatus()
		props["cause"] = randomAzureHealthCause()
	case "StorageRead", "StorageWrite", "StorageDelete":
		props["requestUrl"] = "https://" + randomAzureStorageAccount() + ".blob.core.windows.net/" + randomBlobPath()
		props["userAgentHeader"] = randomUserAgent()
		props["statusCode"] = randomHTTPStatusCode()
		props["serverLatencyMs"] = randomDurationMs()
	}

	return props
}
