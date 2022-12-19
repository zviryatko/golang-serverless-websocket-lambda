package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
)

type MessageData struct {
	Message      string `json:"message"`
	ConnectionID string `json:"connectionId"`
}

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, rawRequest interface{}) (Response, error) {
	reqJson, _ := json.Marshal(rawRequest)
	log.Println("request: ", string(reqJson))
	var request events.APIGatewayWebsocketProxyRequest
	err := json.Unmarshal(reqJson, &request)
	if err != nil {
		log.Println("not a websocket request, err: ")
		return Response{
			StatusCode:      500,
			Body:            "not a websocket request",
			IsBase64Encoded: false,
		}, nil
	}

	switch request.RequestContext.EventType {
	case "CONNECT":
		connect(request)
	case "DISCONNECT":
		disconnect(request)
	case "MESSAGE":
		sendMessage(request.RequestContext.ConnectionID, request.Body)
	}

	return Response{
		StatusCode:      200,
		Body:            "OK",
		IsBase64Encoded: false,
	}, nil
}

func disconnect(request events.APIGatewayWebsocketProxyRequest) {
	client, err := NewDynamoSession()
	if err != nil {
		log.Println(err)
		return
	}
	_, err = client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"connectionId": {
				S: aws.String(request.RequestContext.ConnectionID),
			},
		},
	})
	sendMessage(request.RequestContext.ConnectionID, "disconnected")
}

func connect(request events.APIGatewayWebsocketProxyRequest) {
	client, err := NewDynamoSession()
	if err != nil {
		log.Println(err)
		return
	}
	_, err = client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Item: map[string]*dynamodb.AttributeValue{
			"connectionId": {
				S: aws.String(request.RequestContext.ConnectionID),
			},
		},
	})
	sendMessage(request.RequestContext.ConnectionID, "connected")
}

func sendMessage(connectionId, message string) {
	messageData := &MessageData{
		Message:      message,
		ConnectionID: connectionId,
	}
	jsonData, _ := json.Marshal(messageData)
	client, err := NewDynamoSession()
	if err != nil {
		log.Println(err)
		return
	}
	result, err := client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
	})
	if err != nil {
		log.Println(err)
		return
	}
	apiGatewaySession, err := NewApiGatewaySession()
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range result.Items {
		id := *item["connectionId"].S
		if id == connectionId {
			// skip sending message to the sender.
			continue
		}
		_, err = apiGatewaySession.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(id),
			Data:         jsonData,
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func NewApiGatewaySession() (*apigatewaymanagementapi.ApiGatewayManagementApi, error) {
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(os.Getenv("API_URL")),
		Region:   aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, err
	}
	return apigatewaymanagementapi.New(sess), nil
}

func NewDynamoSession() (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, err
	}
	return dynamodb.New(sess), nil
}

func main() {
	lambda.Start(Handler)
}
