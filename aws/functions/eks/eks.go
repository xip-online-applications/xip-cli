package eks

import (
	"encoding/base64"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	awseks "github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/sts"
)

type Eks struct {
	awsSession *session.Session
	eksClient  *awseks.EKS
}

const (
	RequestPresignParam = 60
	SessionName         = "XIPEKSGetTokenAuth"
	AssumeRoleDuration  = 1800
	ClusterHeaderName   = "x-k8s-aws-id"
	TokenPrefix         = "k8s-aws-v1."
	TokenLifeTime       = 15 * time.Minute
	DateHeaderFormat    = "20060102T150405Z"
	HostRegexp          = `^sts(\.[a-z1-9\-]+)?\.amazonaws\.com(\.cn)?$`
)

type GetTokenOptions struct {
	Region  string
	Cluster string
	RoleArn string
	Session *session.Session
	Sts     *sts.STS
}

func New(awsSession session.Session) Eks {
	eksClient := awseks.New(&awsSession, awsSession.Config)

	return Eks{
		awsSession: &awsSession,
		eksClient:  eksClient,
	}
}

func (e *Eks) GetToken(region string, clusterName string, roleArn string) (string, string, error) {
	// Initial options
	options := &GetTokenOptions{
		Region:  region,
		Cluster: clusterName,
		RoleArn: roleArn,
	}

	// Build new session
	options.Session = e.awsSession.Copy(aws.NewConfig().WithRegion(options.Region).WithSTSRegionalEndpoint(endpoints.RegionalSTSEndpoint))
	options.Sts = sts.New(options.Session)

	return e.GetTokenWithOptions(options)
}

func (e *Eks) GetTokenWithOptions(options *GetTokenOptions) (string, string, error) {
	if options.RoleArn != "" {
		stsClient, err := e.getStsClient(options)

		if err != nil {
			return "", "", err
		}

		options.Sts = &stsClient
	}

	return e.getToken(options)
}

func (e *Eks) getStsClient(options *GetTokenOptions) (sts.STS, error) {
	// duration := int64(AssumeRoleDuration)

	var sessionSetters []func(*stscreds.AssumeRoleProvider)
	sessionSetters = append(sessionSetters, func(provider *stscreds.AssumeRoleProvider) {
		provider.RoleSessionName = SessionName
	})

	creds := stscreds.NewCredentials(options.Session, options.RoleArn, sessionSetters...)

	return *sts.New(options.Session, &aws.Config{Credentials: creds}), nil

	// assumeInput := sts.AssumeRoleInput{
	// 	DurationSeconds: &duration,
	// 	RoleArn:         &options.RoleArn,
	// 	RoleSessionName: &sessionName,
	// }
	//
	// response, err := options.Sts.AssumeRole(&assumeInput)
	// if err != nil {
	// 	return sts.STS{}, err
	// }
	//
	// creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
	// 	AccessKeyID:     *response.Credentials.AccessKeyId,
	// 	SecretAccessKey: *response.Credentials.SecretAccessKey,
	// 	SessionToken:    *response.Credentials.SessionToken,
	// })
	// clientConfig := *aws.NewConfig().
	// 	WithRegion(region).
	// 	WithCredentials(creds)
	//
	// return *sts.New(e.awsSession, &clientConfig), nil
}

func (e *Eks) getToken(options *GetTokenOptions) (string, string, error) {
	// output := &sts.GetCallerIdentityOutput{}
	req, _ := options.Sts.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	req.HTTPRequest.Header.Add(ClusterHeaderName, options.Cluster)

	// req := stsClient.NewRequest(&request.Operation{
	// 	Name:       "GetCallerIdentity",
	// 	HTTPMethod: "GET",
	// 	HTTPPath:   "/",
	// }, &sts.GetCallerIdentityInput{}, output)

	// req.Config.Region = stsClient.Config.Region
	// req.Config.Credentials = stsClient.Config.Credentials

	url, err := req.Presign(RequestPresignParam)
	if err != nil {
		return "", "", err
	}

	tokenExpiration := time.Now().Local().Add(TokenLifeTime - 1*time.Minute)
	return TokenPrefix + base64.RawURLEncoding.EncodeToString([]byte(url)), tokenExpiration.UTC().Format(time.RFC3339), nil

	// encoded := base64.URLEncoding.EncodeToString([]byte(url))
	// encoded = strings.TrimRight(encoded, "=")
	//
	// dur, _ := time.ParseDuration(string(TokenLifeTime))
	// expr := time.Now().Add(dur)
	//
	// return TokenPrefix + encoded, expr.UTC().Format(time.RFC3339), nil
}
