package cli

type Response struct {
	HTTPHeaders    ResponseHeaders `json:"HTTPHeaders"`
	HTTPStatusCode int16           `json:"HTTPStatusCode"`
	RequestId      string          `json:"RequestId"`
	RetryAttempts  int16           `json:"RetryAttempts"`
}

type ResponseHeaders struct {
	ContentLength   string `json:"content-length"`
	ContentType     string `json:"content-type"`
	Date            string `json:"date"`
	AmazonRequestId string `json:"x-amzn-requestid"`
}

type Credentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	Expiration      string `json:"Expiration"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
}

type AssumedRole struct {
	Arn           string `json:"Arn"`
	AssumedRoleId string `json:"AssumedRoleId"`
}
