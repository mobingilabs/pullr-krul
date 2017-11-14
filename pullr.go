package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mobingilabs/mobingi-sdk-go/mobingi/registry/pullr"
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
		TableName: aws.String("MC_IDENTITY"),
	}

	result, err := p.dynamo.GetItem(getInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return "", nil
			}
		}
		return "", err
	}

	githubTokenItem, ok := result.Item["github_token"]
	if !ok || githubTokenItem.S == nil {
		return "", nil
	}

	return *githubTokenItem.S, nil
}

func (p *Pullr) dispatchBuildAction(provider, repository, ref, commit string) error {
	var qc pullr.QueueClient

	type buildData struct {
		Provider   string `json:"provider"`
		Repository string `json:"repository"`
		Ref        string `json:"ref"`
		Commit     string `json:"commit"`
	}

	payload, err := json.Marshal(struct {
		Action string    `json:"action"`
		Data   buildData `json:"data"`
	}{
		Action: "build",
		Data: buildData{
			Provider:   provider,
			Repository: repository,
			Ref:        ref,
			Commit:     commit,
		},
	})

	if err != nil {
		return err
	}

	return qc.Publish(string(payload))
}
