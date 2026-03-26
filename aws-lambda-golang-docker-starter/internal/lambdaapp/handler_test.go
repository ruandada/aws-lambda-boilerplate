package lambdaapp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleRawEvent_HTTPV2(t *testing.T) {
	event := events.APIGatewayV2HTTPRequest{
		Version: "2.0",
		RawPath: "/api/greet/jarry",
		QueryStringParameters: map[string]string{
			"from": "unit-test",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "GET",
				Path:   "/api/greet/jarry",
			},
		},
	}

	raw, _ := json.Marshal(event)
	result, err := HandleRawEvent(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	resp, ok := result.(events.APIGatewayV2HTTPResponse)
	if !ok {
		t.Fatalf("expected APIGatewayV2HTTPResponse, got %T", result)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("expected json body, got error: %v", err)
	}

	if body["message"] != "Hello, jarry!" {
		t.Fatalf("expected greeting message, got %#v", body["message"])
	}

	if body["from"] != "unit-test" {
		t.Fatalf("expected from=unit-test, got %#v", body["from"])
	}
}

func TestHandleRawEvent_HTTPV1(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/api/greet/jarry",
		QueryStringParameters: map[string]string{
			"from": "api-gateway-v1",
		},
	}

	raw, _ := json.Marshal(event)
	result, err := HandleRawEvent(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	resp, ok := result.(events.APIGatewayProxyResponse)
	if !ok {
		t.Fatalf("expected APIGatewayProxyResponse, got %T", result)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("expected json body, got error: %v", err)
	}

	if body["message"] != "Hello, jarry!" {
		t.Fatalf("expected greeting message, got %#v", body["message"])
	}

	if body["from"] != "api-gateway-v1" {
		t.Fatalf("expected from=api-gateway-v1, got %#v", body["from"])
	}
}

func TestHandleRawEvent_SQS(t *testing.T) {
	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId:   "message-1",
				EventSource: "aws:sqs",
				Body:        `{"hello":"world"}`,
			},
		},
	}

	raw, _ := json.Marshal(event)
	result, err := HandleRawEvent(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result for SQS, got %T", result)
	}
}

func TestHandleRawEvent_Unsupported(t *testing.T) {
	result, err := HandleRawEvent(context.Background(), []byte(`{"foo":"bar"}`))
	if err == nil {
		t.Fatalf("expected error for unsupported event, got nil and result=%v", result)
	}
}

func TestHandleRawEvent_InvalidJSON(t *testing.T) {
	result, err := HandleRawEvent(context.Background(), []byte(`{`))
	if err == nil {
		t.Fatalf("expected parsing error, got nil and result=%v", result)
	}
}
