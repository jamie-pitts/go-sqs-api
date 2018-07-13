package main

import (
	"os"
	"net/http"
	"net/url"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

func GetSession() client.ConfigProvider {

	awsRegion := GetEnv("AWS_REGION", "eu-west-1")
	awsProxy := os.Getenv("AWS_PROXY_HOST")
	awsPort := os.Getenv("AWS_PROXY_PORT")

	if awsProxy == "" {
		sess, _ := session.NewSession(
			&aws.Config{
				Region: aws.String(awsRegion)},
		)
		fmt.Printf("AWS session set up using region: %v\n", awsRegion)
		return sess

	} else {
		httpclient := &http.Client{
			Transport: &http.Transport{
				Proxy: func(*http.Request) (*url.URL, error) {
					return url.Parse("http://" + awsProxy + ":" + awsPort)
				},
			},
		}
		sess, _ := session.NewSession(
			&aws.Config{
				Region:     aws.String(awsRegion),
				HTTPClient: httpclient}, )
		fmt.Printf("AWS session set up using region: %v, and proxy: %v:%v\n", awsRegion, awsProxy, awsPort)
		return sess
	}
}

func GetEnv(s string, d string) (string) {
	if os.Getenv(s) == "" { return d } else { return os.Getenv(s) }
}