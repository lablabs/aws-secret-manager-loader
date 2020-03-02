package main

// Use this code snippet in your app.
// If you need more information about configurations or implementing the sample code, visit the AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var (
	region      string
	secretsPath string
)

type Secret struct {
	Key   string
	Value string
}

func getSecret(secretName string) *string {

	//Create a Secrets Manager client
	s, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := secretsmanager.New(s,
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure)
				panic(aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
				panic(aerr.Error())
			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
				panic(aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
				panic(aerr.Error())
			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
				panic(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			panic(err.Error())
		}
		fmt.Println(err.Error())
		panic(err.Error())
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	if result.SecretString != nil {
		return result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))

		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			panic(err)
		}

		secretName := strings.Split(*result.Name, "/")
		f, err := os.Create(secretsPath + secretName[len(secretName)-1])

		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err := f.Write(decodedBinarySecretBytes); err != nil {
			panic(err)
		}
		if err := f.Sync(); err != nil {
			panic(err)
		}
		return nil
	}
}

func listAllSecrets() *secretsmanager.ListSecretsOutput {
	s, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := secretsmanager.New(s,
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.ListSecretsInput{}

	result, err := svc.ListSecrets(input)
	if err != nil {
		fmt.Println(err.Error())
	}
	return result
}

func filterSecrets(targetTags map[string]string) []string {
	allSecrets := listAllSecrets()
	var filteredSecrets []string
	for _, secret := range allSecrets.SecretList {

		// If secret has no tags, skip it
		if len(secret.Tags) == 0 {
			continue
		}

		// Convert tags on resource into map
		resourceTags := make(map[string]string)
		for _, tag := range secret.Tags {
			resourceTags[*tag.Key] = *tag.Value
		}

		// Check if resource has all required tags specified in env
		hasAllTags := true
		for key, value := range targetTags {
			if resourceTags[key] != value {
				hasAllTags = false
				break
			}
		}

		if hasAllTags {
			filteredSecrets = append(filteredSecrets, *secret.Name)
		}
	}

	return filteredSecrets
}

func filterEnvVars(targetPrefix string) map[string]string {

	var result map[string]string

	allVars := os.Environ()
	result = make(map[string]string)
	for _, env := range allVars {

		if strings.HasPrefix(env, targetPrefix) {
			trimmed := strings.TrimPrefix(env, targetPrefix)
			pair := strings.SplitN(trimmed, "=", 2)
			result[pair[0]] = pair[1]
		}
	}

	return result
}

func parseSecrets(secretsNames []string) []string {
	var secrets []string

	for _, secret := range secretsNames {

		var parsedSecret map[string]string

		s := getSecret(secret)
		if s == nil {
			continue
		}

		err := json.Unmarshal([]byte(*s), &parsedSecret)
		if err != nil {
			fmt.Println(err)
		}

		for key, value := range parsedSecret {
			secrets = append(secrets, "export "+key+"='"+value+"'")
		}
	}

	return secrets
}

func main() {

	region = os.Getenv("AWS_REGION")
	secretsPath = os.Getenv("SM_SECRETS_PATH")
	sm_tags := filterEnvVars("SM_TAG_")

	if len(sm_tags) == 0 {
		err := errors.New("No tags for secrets filtering specified")
		panic(err)
	}

	filteredSecretsNames := filterSecrets(sm_tags)
	parsedSecrets := parseSecrets(filteredSecretsNames)

	for _, s := range parsedSecrets {
		fmt.Println(s)
	}
}
