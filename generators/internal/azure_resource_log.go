package internal

import (
	"encoding/json"
	"fmt"
	"math/rand"
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

func NewAzureResourceLogGen(cfg conf.InputConfig) *AzureResourceLogGen {
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
// The Evidence field follows the translator's schema
// (https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log-schema).
type azureAuthorization struct {
	Scope    string         `json:"scope,omitempty"`
	Action   string         `json:"action,omitempty"`
	Evidence *azureEvidence `json:"evidence,omitempty"`
}

// azureEvidence holds role-assignment details inside an authorization block.
type azureEvidence struct {
	Role                string `json:"role,omitempty"`
	RoleAssignmentScope string `json:"roleAssignmentScope,omitempty"`
	RoleAssignmentID    string `json:"roleAssignmentId,omitempty"`
	RoleDefinitionID    string `json:"roleDefinitionId,omitempty"`
	PrincipalID         string `json:"principalId,omitempty"`
	PrincipalType       string `json:"principalType,omitempty"`
}

func buildAzureResourceLog() azureResourceLog {
	category := randomAzureCategory()
	operationName := randomAzureOperationName(category)
	resultType := randomAzureResultType()

	log := azureResourceLog{
		Time:            time.Now().UTC().Format(time.RFC3339Nano),
		ResourceID:      randomAzureResourceID(),
		OperationName:   operationName,
		Category:        category,
		ResultType:      resultType,
		DurationMs:      randomDurationMs(),
		CallerIPAddress: randomIP(),
		CorrelationID:   randomAzureGUID(),
		Level:           randomAzureLogLevel(),
		Location:        randomAzureRegion(),
	}

	// Add result signature for failed operations
	if resultType != "Success" {
		log.ResultSignature = randomAzureErrorCode()
		log.ResultDescription = randomAzureErrorDescription()
	}

	// Administrative and Policy logs carry a full authorization+claims identity;
	// other categories use only a minimal claims block.
	if shouldHaveFullIdentity(category) {
		subscriptionID := randomAzureGUID()

		now := time.Now().Unix()

		log.Identity = &azureIdentity{
			Authorization: &azureAuthorization{
				Scope:  randomAzureResourceID(),
				Action: operationName,
				Evidence: &azureEvidence{
					Role:                randomAzureRole(),
					RoleAssignmentScope: "/subscriptions/" + subscriptionID,
					RoleAssignmentID:    randomAZaz09String(32),
					RoleDefinitionID:    randomAZaz09String(32),
					PrincipalID:         randomAZaz09String(32),
					PrincipalType:       "User",
				},
			},
			Claims: map[string]string{
				"aud":    "https://management.core.windows.net/",
				"iss":    "https://sts.windows.net/" + randomAzureGUID() + "/",
				"iat":    strconv.FormatInt(now-int64(rand.Intn(300)), 10), // issued up to 5 min ago
				"nbf":    strconv.FormatInt(now, 10),                       // not valid before now
				"exp":    strconv.FormatInt(now+int64(3600), 10),           // expires in 1 hour
				"name":   randomAzureUserName(),
				"ipaddr": randomIP(),
			},
		}
	} else {
		log.Identity = &azureIdentity{
			Claims: map[string]string{
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/spn": randomAzureServicePrincipal(category),
			},
		}
	}

	// Add category-specific properties
	log.Properties = generateAzureProperties(category, operationName)

	return log
}

// shouldHaveFullIdentity returns true for categories that carry a full
// identity object (authorization + claims) in real Azure activity logs.
func shouldHaveFullIdentity(category string) bool {
	return category == "Administrative" || category == "Policy"
}

// generateAzureProperties returns a properties map whose keys match the
// schemas expected by the opentelemetry-collector-contrib azurelogs translator.
// See: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/translator/azurelogs
func generateAzureProperties(category, operationName string) map[string]interface{} {
	props := make(map[string]interface{})

	switch category {
	case "Administrative":
		// administrativeLogProperties: entity, message, hierarchy (+ eventCategory)
		resourceID := randomAzureResourceID()
		tenantID := randomAzureGUID()
		subscriptionID := randomAzureGUID()
		props["eventCategory"] = "Administrative"
		props["entity"] = resourceID
		props["message"] = operationName
		props["hierarchy"] = tenantID + "/" + subscriptionID

	case "Security":
		// securityLogProperties: accountLogonId, commandLine, domainName,
		// parentProcess, parentProcessid, processId, processName,
		// userName, UserSID, ActionTaken, Severity
		processName := fmt.Sprintf("c:\\windows\\system32\\%s.exe", randomAZaz09String(6))
		props["eventCategory"] = "Security"
		props["accountLogonId"] = fmt.Sprintf("0x%s", randomAZaz09String(4))
		props["commandLine"] = processName
		props["domainName"] = randomAZaz09String(6)
		props["parentProcess"] = "explorer.exe"
		props["parentProcess id"] = fmt.Sprintf("%d", rand.Intn(9000)+1000)
		props["processId"] = fmt.Sprintf("%d", rand.Intn(9000)+1000)
		props["processName"] = processName
		props["userName"] = fmt.Sprintf("user%s", randomAZaz09String(4))
		props["UserSID"] = fmt.Sprintf("S-1-5-21-%d-%d-%d", rand.Intn(9999999), rand.Intn(9999999), rand.Intn(9999999))
		props["ActionTaken"] = randomAzureSecurityActionTaken()
		props["Severity"] = randomAzureSecuritySeverity()

	case "ServiceHealth":
		// serviceHealthLogProperties: title, service, region, communication,
		// communicationId, incidentType, trackingId, impactStartTime,
		// impactMitigationTime, impactedServices
		service := randomAzureServiceName()
		region := randomAzureRegion()
		trackingID := randomAZaz09String(8)
		impactStart := time.Now().UTC().Add(-time.Duration(rand.Intn(72)) * time.Hour)
		props["title"] = fmt.Sprintf("Service issue with %s in %s", service, region)
		props["service"] = service
		props["region"] = region
		props["communication"] = fmt.Sprintf("We are aware of an issue with %s in %s and are actively investigating.", service, region)
		props["communicationId"] = randomAZaz09String(12)
		props["incidentType"] = randomAzureIncidentType()
		props["trackingId"] = trackingID
		props["impactStartTime"] = impactStart.Format(time.RFC3339Nano)
		props["impactMitigationTime"] = impactStart.Add(time.Duration(rand.Intn(24)+1) * time.Hour).Format(time.RFC3339Nano)
		props["impactedServices"] = fmt.Sprintf(
			`[{"ImpactedRegions":[{"RegionName":"%s"}],"ServiceName":"%s"}]`,
			region, service,
		)
		props["stage"] = "Active"
		props["isHIR"] = false
		props["IsSynthetic"] = "False"

	case "ResourceHealth":
		// resourceHealthLogProperties: title, details, currentHealthStatus,
		// previousHealthStatus, type, cause
		currentStatus := randomAzureHealthStatus()
		previousStatus := randomAzureHealthStatus()
		props["title"] = currentStatus
		props["details"] = fmt.Sprintf("Resource transitioned from %s to %s", previousStatus, currentStatus)
		props["currentHealthStatus"] = currentStatus
		props["previousHealthStatus"] = previousStatus
		props["type"] = randomAzureHealthType()
		props["cause"] = randomAzureHealthCause()

	case "Alert":
		// alertLogProperties: RuleUri, RuleName, RuleDescription, Threshold,
		// WindowSizeInMinutes, Aggregation, Operator, MetricName, MetricUnit
		ruleName := fmt.Sprintf("alert-%s", randomAZaz09String(6))
		resourceID := randomAzureResourceID()
		props["RuleUri"] = resourceID + "/providers/microsoft.insights/alertrules/" + ruleName
		props["RuleName"] = ruleName
		props["RuleDescription"] = fmt.Sprintf("Alert rule for %s", randomAzureAlertMetricName())
		props["Threshold"] = fmt.Sprintf("%d", rand.Intn(99000)+1000)
		props["WindowSizeInMinutes"] = fmt.Sprintf("%d", []int{5, 10, 15, 30, 60}[rand.Intn(5)])
		props["Aggregation"] = randomAzureAlertAggregation()
		props["Operator"] = randomAzureAlertOperator()
		props["MetricName"] = randomAzureAlertMetricName()
		props["MetricUnit"] = "Count"

	case "Recommendation":
		// recommendationLogProperties: recommendationSchemaVersion,
		// recommendationCategory, recommendationImpact, recommendationName,
		// recommendationResourceLink, recommendationType
		recommendationType := randomAzureRecommendationType()
		resourceID := randomAzureResourceID()
		props["recommendationSchemaVersion"] = "1.0"
		props["recommendationCategory"] = randomAzureRecommendationCategory()
		props["recommendationImpact"] = randomAzureRecommendationImpact()
		props["recommendationName"] = randomAzureRecommendationName()
		props["recommendationResourceLink"] = fmt.Sprintf(
			"https://portal.azure.com/#blade/Microsoft_Azure_Expert/RecommendationListBlade/recommendationTypeId/%s/resourceId/%s",
			recommendationType, resourceID,
		)
		props["recommendationType"] = recommendationType

	case "Policy":
		// policyLogProperties: isComplianceCheck, resourceLocation, ancestors,
		// policies (JSON string), eventCategory, entity, message, hierarchy
		subscriptionID := randomAzureGUID()
		resourceID := randomAzureResourceID()
		policyDefID := randomAzurePolicyDefinitionID()
		policyJSON := fmt.Sprintf(
			`[{"policyDefinitionId":"%s","policyDefinitionEffect":"AuditIfNotExists","policyAssignmentScope":"/subscriptions/%s"}]`,
			policyDefID, subscriptionID,
		)
		props["isComplianceCheck"] = "False"
		props["resourceLocation"] = randomAzureRegion()
		props["ancestors"] = subscriptionID
		props["policies"] = policyJSON
		props["eventCategory"] = "Policy"
		props["entity"] = resourceID
		props["message"] = operationName
		props["hierarchy"] = ""

	case "Autoscale":
		// autoscaleLogProperties: Description, ResourceName, OldInstancesCount,
		// NewInstancesCount, LastScaleActionTime
		resourceName := randomAzureResourceID()
		oldCount := rand.Intn(8) + 2
		newCount := oldCount + []int{-1, 1}[rand.Intn(2)]
		if newCount < 1 {
			newCount = 1
		}
		props["Description"] = fmt.Sprintf(
			"The autoscale engine attempting to scale resource '%s' from %d instances count to %d instances count.",
			resourceName, oldCount, newCount,
		)
		props["ResourceName"] = resourceName
		props["OldInstancesCount"] = fmt.Sprintf("%d", oldCount)
		props["NewInstancesCount"] = fmt.Sprintf("%d", newCount)
		props["LastScaleActionTime"] = time.Now().UTC().Format(time.RFC1123)
	}

	return props
}
