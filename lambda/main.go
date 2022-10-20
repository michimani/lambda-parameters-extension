package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Response struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

type Payload struct {
	UseExtension     bool `json:"useExtension"`
	Count            int  `json:"count"`
	ParameterVersion *int `json:"parameterVersion,omitempty"`
}

const paramKey = "/test/lambda-parameters-extension"

var env string
var ssmClient *ssm.Client

func handleRequest(ctx context.Context, payload Payload) (Response, error) {
	initLog(payload)
	defer log.Printf("end handler at %s", env)

	var f func(k string, v int) (string, error)
	if payload.UseExtension {
		f = getValueByUsingExtension
	} else {
		f = getValueByCallingParameterStoreAPI
	}

	parameterVersion := 1
	if payload.ParameterVersion != nil {
		parameterVersion = *payload.ParameterVersion
	}

	// get a value
	for i := 0; i < payload.Count; i++ {
		if v, err := f(paramKey, parameterVersion); err != nil {
			return Response{
				Message:    fmt.Sprintf("Failed to get a value for key:%s. err:%v", paramKey, err),
				StatusCode: http.StatusInternalServerError,
			}, err
		} else {
			log.Printf("[%d] extension:%v value:%s key:%s", i, payload.UseExtension, v, paramKey)
		}
	}

	return Response{
		StatusCode: http.StatusOK,
	}, nil
}

// for cold start
func init() {
	log.Println("cold start")

	region := os.Getenv("AWS_DEFAULT_REGION")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load config for using SDK. err:%v", err)
	}

	ssmClient = ssm.NewFromConfig(cfg)
	log.Println("initialized SSM client")
}

// for every invocation
func initLog(payload Payload) {
	env = os.Getenv("ENV")
	log.Printf("start handler at %s", env)
	log.Printf("UseExtension: %v, Count: %d", payload.UseExtension, payload.Count)
}

// Get a value from Parameter Store directory.
func getValueByCallingParameterStoreAPI(key string, version int) (string, error) {
	in := ssm.GetParameterInput{
		Name: aws.String(fmt.Sprintf("%s:%d", key, version)),
	}
	out, err := ssmClient.GetParameter(context.Background(), &in)
	if err != nil {
		return "", err
	}

	return *out.Parameter.Value, nil
}

const (
	// Endpoint for getting parameter by Parameters and Secrets Lambda Extension.
	exGetParameterEndpoint = "http://localhost:2773/systemsmanager/parameters/get"

	// Header key of secret token
	secretTokenHeaderKey = "X-Aws-Parameters-Secrets-Token"

	// Query parameter key
	queryParameterKeyForName    = "name"
	queryParameterKeyForVersion = "version"
)

// Struct of response from AWSParametersAndSecretsLambdaExtension API
type resultFromExtension struct {
	Parameter struct {
		ARN              string
		DateType         string
		LastModifiedDate time.Time
		Name             string
		Selector         string
		SourceResult     *string
		Type             string
		Value            string
		Version          int
	}
	ResultMetadata any
}

// Get a value using Parameters and Secrets Lambda Extension.
func getValueByUsingExtension(key string, version int) (string, error) {
	// Get a value from extension
	// https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html
	query := url.Values{}
	query.Add(queryParameterKeyForName, key)
	query.Add(queryParameterKeyForVersion, fmt.Sprintf("%d", version))
	queryStr := query.Encode()

	url := fmt.Sprintf("%s?%s", exGetParameterEndpoint, queryStr)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return "", err
	}

	// set X-Aws-Parameters-Secrets-Token header
	req.Header.Add(secretTokenHeaderKey, os.Getenv("AWS_SESSION_TOKEN"))

	// call Extension API
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return "", err
	}
	bodyString := buf.String()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to get parameter by using extension. statusCode:%d body:%s", res.StatusCode, bodyString)
	}

	exRes := resultFromExtension{}
	if err := json.Unmarshal([]byte(bodyString), &exRes); err != nil {
		return "", err
	}

	return exRes.Parameter.Value, nil
}

func main() {
	runtime.Start(handleRequest)
}
