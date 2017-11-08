package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func getUsernameByRepository(provider, repository string) (string, error) {
	cfg := &aws.Config{}
	sess := session.Must(session.NewSession(cfg))
	dynamo := dynamodb.New(sess)

	repositoryPair := fmt.Sprintf("%s:%s", provider, repository)
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"repository": {S: aws.String(repositoryPair)},
		},
		TableName: aws.String("PULLR_REPOS"),
	}

	result, err := dynamo.GetItem(getInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return "", nil
			}
			return "", aerr
		}
		return "", err
	}

	return *result.Item["username"].S, nil
}

func getGithubTokenByUsername(username string) (string, error) {
	cfg := &aws.Config{}
	sess := session.Must(session.NewSession(cfg))
	dynamo := dynamodb.New(sess)

	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"username": {S: aws.String(username)},
		},
	}

	result, err := dynamo.GetItem(getInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return "", nil
			}
			return "", aerr
		}
		return "", err
	}

	githubTokenItem, ok := result.Item["github_token"]
	if !ok || githubTokenItem.S == nil {
		return "", nil
	}

	return *githubTokenItem.S, nil
}
