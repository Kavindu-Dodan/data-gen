package main

type awsMetric struct {
	MetricStreamName string `json:"metric_stream_name"`
	AccountId        string `json:"account_id"`
	Region           string `json:"region"`
	Namespace        string `json:"namespace"`
	MetricName       string `json:"metric_name"`
	Dimensions       struct {
		InstanceId string `json:"InstanceId"`
	} `json:"dimensions"`
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Average float64 `json:"Average"`
		Minimum float64 `json:"Minimum"`
		Maximum float64 `json:"Maximum"`
	} `json:"value"`
	Unit string `json:"unit"`
}
