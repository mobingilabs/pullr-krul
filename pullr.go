package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Pullr struct {
	awsSess *session.Session
	dynamo  *dynamodb.DynamoDB
}

func NewPullr(awsSession *session.Session) Pullr {
	dynamo := dynamodb.New(awsSession)
	return Pullr{awsSess: awsSession, dynamo: dynamo}
}

func (p *Pullr) getUsernameByRepository(provider, repository string) (string, error) {
	repositoryPair := fmt.Sprintf("%s:%s", provider, repository)
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"repository": {S: aws.String(repositoryPair)},
		},
		TableName: aws.String("PULLR_REPOS"),
	}

	result, err := p.dynamo.GetItem(getInput)
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

func (p *Pullr) getGithubTokenByUsername(username string) (string, error) {
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"username": {S: aws.String(username)},
		},
		TableName: aws.String("PULLR_REPOS"),
	}

	result, err := p.dynamo.GetItem(getInput)
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
