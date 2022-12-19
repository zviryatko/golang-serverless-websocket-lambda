# Example of Serverless + AWS Lambda + Golang + Websockets + DynamoDB

This is an example of a Serverless project that uses AWS Lambda and Golang to create a Websocket through AWS API Gateway and stores sessions in AWS DynamoDB.

## Requirements

- [Serverless](https://serverless.com/)
- [Golang](https://golang.org/)
- [AWS Account](https://aws.amazon.com/)
- [AWS CLI](https://aws.amazon.com/cli/)

## Deploy

1. Clone this repository
2. Run `make deploy` to deploy AWS Stack

## Usage

Open from different terminals `wscat -c wss://<API Gateway URL>/dev`
Or connect to Websocket from different browser tabs.