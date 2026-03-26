package lambdaapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func HandleRawEvent(_ context.Context, raw json.RawMessage) (any, error) {
	eventType, err := detectEventType(raw)
	if err != nil {
		return nil, err
	}

	switch eventType {
	case EventTypeHTTPV2:
		var event events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		return handleHTTPv2(event), nil
	case EventTypeHTTPV1:
		var event events.APIGatewayProxyRequest
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		return handleHTTPv1(event), nil
	case EventTypeSQS:
		var event events.SQSEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		for _, record := range event.Records {
			fmt.Printf("SQS message received: %s\n", record.MessageId)
		}
		return nil, nil
	default:
		return nil, errors.New("unsupported event type")
	}
}

func handleHTTPv2(event events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	method := strings.ToUpper(event.RequestContext.HTTP.Method)
	path := event.RawPath
	if strings.TrimSpace(path) == "" {
		path = event.RequestContext.HTTP.Path
	}

	if method == httpMethodGet && path == "/" {
		return jsonResponseV2(200, map[string]any{
			"message": "Hello World",
		})
	}

	if method == httpMethodGet && strings.HasPrefix(path, "/api/greet/") {
		name := strings.TrimPrefix(path, "/api/greet/")
		if strings.TrimSpace(name) == "" {
			return jsonResponseV2(404, map[string]any{"message": "Not Found"})
		}

		from := strings.TrimSpace(event.QueryStringParameters["from"])
		if from == "" {
			from = "starter"
		}

		return jsonResponseV2(200, map[string]any{
			"message": fmt.Sprintf("Hello, %s!", name),
			"from":    from,
		})
	}

	return jsonResponseV2(404, map[string]any{"message": "Not Found"})
}

func handleHTTPv1(event events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	method := strings.ToUpper(event.HTTPMethod)
	path := event.Path

	if method == httpMethodGet && path == "/" {
		return jsonResponseV1(200, map[string]any{
			"message": "Hello World",
		})
	}

	if method == httpMethodGet && strings.HasPrefix(path, "/api/greet/") {
		name := strings.TrimPrefix(path, "/api/greet/")
		if strings.TrimSpace(name) == "" {
			return jsonResponseV1(404, map[string]any{"message": "Not Found"})
		}

		from := strings.TrimSpace(event.QueryStringParameters["from"])
		if from == "" {
			from = "starter"
		}

		return jsonResponseV1(200, map[string]any{
			"message": fmt.Sprintf("Hello, %s!", name),
			"from":    from,
		})
	}

	return jsonResponseV1(404, map[string]any{"message": "Not Found"})
}

func jsonResponseV2(statusCode int, payload any) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(payload)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

func jsonResponseV1(statusCode int, payload any) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(payload)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

const httpMethodGet = "GET"
