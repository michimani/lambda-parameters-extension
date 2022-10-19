package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	UseExtension bool `json:"useExtension"`
	Count        int  `json:"count"`
}

const paramKey = "/test/lambda-parameters-extension"

var env string
var ssmClient *ssm.Client

func handleRequest(ctx context.Context, payload Payload) (Response, error) {
	initLog(payload)
	defer log.Printf("end handler at %s", env)

	var f func(k string) (string, error)
	if payload.UseExtension {
		f = getValueByUsingExtension
	} else {
		f = getValueByCallingParameterStoreAPI
	}

	// get a value
	for i := 0; i < payload.Count; i++ {
		if v, err := f(paramKey); err != nil {
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
func getValueByCallingParameterStoreAPI(key string) (string, error) {
	in := ssm.GetParameterInput{
		Name: aws.String(key),
	}
	out, err := ssmClient.GetParameter(context.Background(), &in)
	if err != nil {
		return "", err
	}

	return *out.Parameter.Value, nil
}

// Get a value using Parameters and Secrets Lambda Extension.
// NOTE:
// Performs as same as getValueByCallingParameterStoreAPI when at local
// because that the extension does not work at local.
func getValueByUsingExtension(key string) (string, error) {
	if env == "local" {
		return getValueByCallingParameterStoreAPI(key)
	}

	value := ""

	return value, nil
}

func main() {
	runtime.Start(handleRequest)
}
