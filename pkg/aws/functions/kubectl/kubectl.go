package kubectl

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"

	"xip/aws/functions/config/config"
)

type Kubectl struct {
	EksClient *eks.EKS
	AwsConfig config.Config
}

func New(AwsSession *session.Session, AwsConfig config.Config) Kubectl {
	eksClient := eks.New(AwsSession)

	return Kubectl{
		EksClient: eksClient,
		AwsConfig: AwsConfig,
	}
}

func (k *Kubectl) RegisterProfile(clusterName string, roleArn string, profile string, namespace string, alias string) error {
	var (
		err             error
		cluster         eks.Cluster
		certificatePath string
		clusterId       string
		userId          string
		contextId       string
	)

	// Retrieve the credentials of the profile
	profileCredentials, err := k.AwsConfig.GetProfile(profile)
	if err != nil {
		return fmt.Errorf("could not retrieve credentials for profile " + profile)
	}

	// Retrieve cluster information and write certificate
	if cluster, err = k.GetEksCluster(clusterName); err != nil {
		return err
	}
	if certificatePath, err = k.WriteCertificate(cluster); err != nil {
		return err
	}
	defer k.RemoveCertificate(certificatePath)

	// Register the cluster
	if clusterId, err = k.RegisterCluster(cluster, certificatePath); err != nil {
		return err
	}

	// Register the user
	if userId, err = k.RegisterUser(cluster, profileCredentials, roleArn); err != nil {
		return err
	}

	// Register context
	if contextId, err = k.RegisterContext(clusterId, userId, namespace, alias); err != nil {
		return err
	}

	// Set current active context
	if err = k.SetActiveContext(contextId); err != nil {
		return err
	}

	return nil
}

func (k *Kubectl) GetEksCluster(name string) (eks.Cluster, error) {
	input := eks.DescribeClusterInput{Name: &name}
	cluster, err := k.EksClient.DescribeCluster(&input)

	if err != nil {
		return eks.Cluster{}, err
	}

	return *cluster.Cluster, nil
}

func (k *Kubectl) WriteCertificate(cluster eks.Cluster) (string, error) {
	// Construct path
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	path := filepath.FromSlash(usr.HomeDir + "/.xip/" + *cluster.Name + ".pem")

	certificate, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return "", err
	}

	// Write credentials
	err = ioutil.WriteFile(path, []byte(certificate), 0644)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (k *Kubectl) RemoveCertificate(path string) {
	_ = os.Remove(path)
}

func (k *Kubectl) RegisterCluster(cluster eks.Cluster, certificatePath string) (string, error) {
	// Set cluster informtion
	err := exec.Command(
		"kubectl", "config", "set-cluster", *cluster.Arn,
		"--certificate-authority", certificatePath,
		"--server", *cluster.Endpoint, "--embed-certs").Run()

	if err != nil {
		return "", err
	}

	return *cluster.Arn, nil
}

func (k *Kubectl) RegisterUser(cluster eks.Cluster, profile config.EntryConfig, roleArn string) (string, error) {
	credentialsArgs := []string{
		"aws",
		"eks-token",
		"--cluster-name",
		*cluster.Name,
	}

	if len(profile.Name) > 0 {
		credentialsArgs = append(credentialsArgs, "--profile", profile.Name)
	}

	if len(roleArn) > 0 {
		credentialsArgs = append(credentialsArgs, "--role-arn", roleArn)
	}

	commandArgs := []string{
		"config", "set-credentials", *cluster.Arn + "_" + roleArn,
		"--exec-api-version", AuthApiVersion,
		"--exec-command", "x-ip",
	}

	for _, element := range credentialsArgs {
		commandArgs = append(commandArgs, "--exec-arg", element)
	}

	// Set user information
	err := exec.Command("kubectl", commandArgs...).Run()
	if err != nil {
		return "", err
	}

	return *cluster.Arn + "_" + roleArn, nil
}

func (k *Kubectl) RegisterContext(clusterId string, userId string, namespace string, alias string) (string, error) {
	name := alias
	if len(alias) == 0 {
		name = userId

		if len(namespace) > 0 {
			name = name + "_" + namespace
		}
	}

	// Build arguments
	arguments := []string{
		"config", "set-context", name,
		"--cluster", clusterId,
		"--user", userId,
	}

	if len(namespace) > 0 {
		arguments = append(arguments, "--namespace", namespace)
	}

	// Set context
	err := exec.Command("kubectl", arguments...).Run()
	if err != nil {
		return "", err
	}

	return name, nil
}

func (k *Kubectl) SetActiveContext(contextId string) error {
	// Set cluster informtion
	err := exec.Command("kubectl", "config", "use-context", contextId).Run()
	if err != nil {
		return err
	}

	return nil
}
