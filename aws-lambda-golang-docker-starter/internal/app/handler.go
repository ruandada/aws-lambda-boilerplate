package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	awsevents "github.com/aws/aws-lambda-go/events"
	"github.com/ruandada/aws-lambda-boilerplate/aws-lambda-golang-docker-starter/internal/events"
)

var sharedMux = NewMux()

func HandleRawEvent(_ context.Context, raw json.RawMessage) (any, error) {
	envelope, err := events.ParseEnvelope(raw)
	if err != nil {
		return nil, err
	}

	if event, ok := events.IsHTTPV2(raw, envelope); ok {
		return handleHTTPv2(event), nil
	}

	if event, ok := events.IsHTTPV1(raw, envelope); ok {
		return handleHTTPv1(event), nil
	}

	if event, ok := events.IsSQS(raw, envelope); ok {
		for _, record := range event.Records {
			fmt.Printf("SQS message received: %s\n", record.MessageId)
		}
		return nil, nil
	}

	return nil, errors.New("unsupported event type")
}

func handleHTTPv2(event awsevents.APIGatewayV2HTTPRequest) awsevents.APIGatewayV2HTTPResponse {
	rawQuery := event.RawQueryString
	if rawQuery == "" {
		qs := url.Values{}
		for k, v := range event.QueryStringParameters {
			qs.Set(k, v)
		}
		rawQuery = qs.Encode()
	}

	path := event.RawPath
	if strings.TrimSpace(path) == "" {
		path = event.RequestContext.HTTP.Path
	}

	req := buildHTTPRequest(
		event.RequestContext.HTTP.Method,
		path,
		rawQuery,
		event.Headers,
		event.Body,
	)

	rec := httptest.NewRecorder()
	sharedMux.ServeHTTP(rec, req)

	return awsevents.APIGatewayV2HTTPResponse{
		StatusCode: rec.Code,
		Headers:    flattenHeaders(rec.Header()),
		Body:       rec.Body.String(),
	}
}

func handleHTTPv1(event awsevents.APIGatewayProxyRequest) awsevents.APIGatewayProxyResponse {
	qs := url.Values{}
	for k, v := range event.QueryStringParameters {
		qs.Set(k, v)
	}

	req := buildHTTPRequest(
		event.HTTPMethod,
		event.Path,
		qs.Encode(),
		event.Headers,
		event.Body,
	)

	rec := httptest.NewRecorder()
	sharedMux.ServeHTTP(rec, req)

	return awsevents.APIGatewayProxyResponse{
		StatusCode: rec.Code,
		Headers:    flattenHeaders(rec.Header()),
		Body:       rec.Body.String(),
	}
}

func buildHTTPRequest(method, path, rawQuery string, headers map[string]string, body string) *http.Request {
	method = strings.ToUpper(method)
	if strings.TrimSpace(path) == "" {
		path = "/"
	}

	u := &url.URL{Path: path, RawQuery: rawQuery}
	req := httptest.NewRequest(method, u.String(), strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, vs := range h {
		out[k] = strings.Join(vs, ", ")
	}
	return out
}
