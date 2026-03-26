package lambdaapp

import (
	"encoding/json"
	"strings"

	awsevents "github.com/aws/aws-lambda-go/events"
)

type EventType string

const (
	EventTypeHTTPV2                 EventType = "http-v2"
	EventTypeLambdaFunctionURL      EventType = "lambda-function-url"
	EventTypeALB                    EventType = "alb"
	EventTypeHTTPV1                 EventType = "http-v1"
	EventTypeSQS                    EventType = "sqs"
	EventTypeSNS                    EventType = "sns"
	EventTypeS3                     EventType = "s3"
	EventTypeSES                    EventType = "ses"
	EventTypeDynamoDBStream         EventType = "dynamodb-stream"
	EventTypeKinesisStream          EventType = "kinesis-stream"
	EventTypeCloudWatchLogs         EventType = "cloudwatch-logs"
	EventTypeScheduled              EventType = "scheduled"
	EventTypeCodeBuildState         EventType = "codebuild-state"
	EventTypeCodeCommit             EventType = "codecommit"
	EventTypeEventBridge            EventType = "eventbridge"
	EventTypeCodePipeline           EventType = "codepipeline"
	EventTypeCloudWatchAlarm        EventType = "cloudwatch-alarm"
	EventTypeConnectContactFlow     EventType = "connect-contact-flow"
	EventTypeCloudFrontRequest      EventType = "cloudfront-request"
	EventTypeCloudFrontResponse     EventType = "cloudfront-response"
	EventTypeGuardDutyNotification  EventType = "guardduty-notification"
	EventTypeS3Batch                EventType = "s3-batch"
	EventTypeS3Notification         EventType = "s3-notification"
	EventTypeLex                    EventType = "lex"
	EventTypeLexV2                  EventType = "lex-v2"
	EventTypeAppSyncResolver        EventType = "appsync-resolver"
	EventTypeAmplifyResolver        EventType = "amplify-resolver"
	EventTypeAPIGatewayAuthorizer   EventType = "api-gateway-authorizer"
	EventTypeAutoScalingScaleIn     EventType = "autoscaling-scale-in"
	EventTypeCloudFormationCustom   EventType = "cloudformation-custom-resource"
	EventTypeCdkCustomResource      EventType = "cdk-custom-resource"
	EventTypeIoT                    EventType = "iot"
	EventTypeIoTCustomAuthorizer    EventType = "iot-custom-authorizer"
	EventTypeFirehoseTransformation EventType = "firehose-transformation"
	EventTypeSecretsManagerRotation EventType = "secrets-manager-rotation"
	EventTypeTransferFamilyAuth     EventType = "transfer-family-authorizer"
	EventTypeMSK                    EventType = "msk"
	EventTypeSelfManagedKafka       EventType = "self-managed-kafka"
	EventTypeUnknown                EventType = "unknown"
)

func detectEventType(raw json.RawMessage) (EventType, error) {
	// Parse once up front so malformed JSON always returns a hard error.
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return EventTypeUnknown, err
	}

	// IoT events can be scalar payloads (string/number).
	switch payload.(type) {
	case string, float64:
		return EventTypeIoT, nil
	}

	envelope, ok := payload.(map[string]any)
	if !ok {
		return EventTypeUnknown, nil
	}

	switch {
	case matches(raw, func(event awsevents.APIGatewayV2HTTPRequest) bool {
		return event.Version == "2.0" && strings.TrimSpace(event.RequestContext.HTTP.Method) != ""
	}):
		if domainName := strings.TrimSpace(toString(envelope["requestContext"], "domainName")); strings.Contains(domainName, "lambda-url") {
			return EventTypeLambdaFunctionURL, nil
		}
		return EventTypeHTTPV2, nil
	case nestedObject(envelope, "requestContext", "elb") != nil:
		return EventTypeALB, nil
	case matches(raw, func(event awsevents.APIGatewayProxyRequest) bool {
		return strings.TrimSpace(event.HTTPMethod) != ""
	}):
		return EventTypeHTTPV1, nil
	case matches(raw, func(event awsevents.SQSEvent) bool {
		return len(event.Records) > 0 && event.Records[0].EventSource == "aws:sqs"
	}):
		return EventTypeSQS, nil
	case matches(raw, func(event awsevents.SNSEvent) bool {
		return len(event.Records) > 0 && event.Records[0].EventSource == "aws:sns"
	}):
		return EventTypeSNS, nil
	case matches(raw, func(event awsevents.S3Event) bool {
		return len(event.Records) > 0 && event.Records[0].EventSource == "aws:s3"
	}):
		return EventTypeS3, nil
	case matches(raw, func(event awsevents.SimpleEmailEvent) bool {
		return len(event.Records) > 0 && event.Records[0].EventSource == "aws:ses"
	}):
		return EventTypeSES, nil
	case matches(raw, func(event awsevents.DynamoDBEvent) bool {
		return len(event.Records) > 0 && event.Records[0].EventSource == "aws:dynamodb"
	}):
		return EventTypeDynamoDBStream, nil
	case matches(raw, func(event awsevents.KinesisEvent) bool {
		return len(event.Records) > 0 && event.Records[0].EventSource == "aws:kinesis"
	}):
		return EventTypeKinesisStream, nil
	case matches(raw, func(event awsevents.CloudwatchLogsEvent) bool {
		return strings.TrimSpace(event.AWSLogs.Data) != ""
	}):
		return EventTypeCloudWatchLogs, nil
	case matches(raw, func(event awsevents.CloudWatchEvent) bool {
		return event.Source == "aws.events" && event.DetailType == "Scheduled Event"
	}):
		return EventTypeScheduled, nil
	case matches(raw, func(event awsevents.CodeBuildEvent) bool {
		return event.Source == "aws.codebuild" && event.DetailType == "CodeBuild Build State Change"
	}):
		return EventTypeCodeBuildState, nil
	case hasCodeCommitRecord(envelope):
		return EventTypeCodeCommit, nil
	case hasString(envelope, "alarmArn") && isObject(envelope["alarmData"]):
		return EventTypeCloudWatchAlarm, nil
	case envelope["Name"] == "ContactFlowEvent" && isObject(nestedObject(envelope, "Details", "ContactData")):
		return EventTypeConnectContactFlow, nil
	case isCloudFrontRequestEvent(envelope):
		return EventTypeCloudFrontRequest, nil
	case isCloudFrontResponseEvent(envelope):
		return EventTypeCloudFrontResponse, nil
	case matches(raw, func(event awsevents.CloudWatchEvent) bool {
		return event.Source == "aws.guardduty"
	}):
		return EventTypeGuardDutyNotification, nil
	case hasString(envelope, "invocationSchemaVersion") && isArray(envelope["tasks"]):
		return EventTypeS3Batch, nil
	case matches(raw, func(event awsevents.CloudWatchEvent) bool {
		return event.Source == "aws.s3" && strings.TrimSpace(event.DetailType) != ""
	}):
		return EventTypeS3Notification, nil
	case hasString(envelope, "messageVersion") && isObject(envelope["currentIntent"]) && isObject(envelope["bot"]):
		return EventTypeLex, nil
	case hasString(envelope, "messageVersion") && isObject(envelope["bot"]) && hasString(envelope, "sessionId"):
		return EventTypeLexV2, nil
	case hasString(nestedObject(envelope, "info"), "fieldName") && isObject(envelope["request"]):
		return EventTypeAppSyncResolver, nil
	case hasString(envelope, "typeName") && hasString(envelope, "fieldName") && isObject(envelope["request"]):
		return EventTypeAmplifyResolver, nil
	case isAPIGatewayAuthorizerEvent(envelope):
		return EventTypeAPIGatewayAuthorizer, nil
	case hasString(envelope, "AutoScalingGroupARN") && isArray(envelope["CapacityToTerminate"]):
		return EventTypeAutoScalingScaleIn, nil
	case isCloudFormationCustomResourceEvent(envelope):
		// CDK custom resource event shares the same envelope shape.
		return EventTypeCloudFormationCustom, nil
	case isObject(envelope["protocolData"]) && isObject(envelope["connectionMetadata"]):
		return EventTypeIoTCustomAuthorizer, nil
	case matches(raw, func(event awsevents.CloudWatchEvent) bool {
		return strings.TrimSpace(event.Source) != "" && strings.TrimSpace(event.DetailType) != ""
	}):
		return EventTypeEventBridge, nil
	case matches(raw, func(event awsevents.CodePipelineEvent) bool {
		return strings.TrimSpace(event.CodePipelineJob.ID) != ""
	}):
		return EventTypeCodePipeline, nil
	case matches(raw, func(event awsevents.KinesisFirehoseEvent) bool {
		return strings.TrimSpace(event.InvocationID) != "" && len(event.Records) > 0
	}):
		return EventTypeFirehoseTransformation, nil
	case matches(raw, func(event awsevents.SecretsManagerSecretRotationEvent) bool {
		return strings.TrimSpace(event.Step) != "" &&
			strings.TrimSpace(event.SecretID) != "" &&
			strings.TrimSpace(event.ClientRequestToken) != ""
	}):
		return EventTypeSecretsManagerRotation, nil
	case hasString(envelope, "username") && hasString(envelope, "serverId") && hasString(envelope, "sourceIp"):
		return EventTypeTransferFamilyAuth, nil
	case matches(raw, func(event awsevents.KafkaEvent) bool {
		return event.EventSource == "aws:kafka" && len(event.Records) > 0
	}):
		return EventTypeMSK, nil
	case matches(raw, func(event awsevents.KafkaEvent) bool {
		return event.EventSource == "SelfManagedKafka" && len(event.Records) > 0
	}):
		return EventTypeSelfManagedKafka, nil
	}

	return EventTypeUnknown, nil
}

func matches[T any](raw json.RawMessage, matcher func(event T) bool) bool {
	var event T
	if err := json.Unmarshal(raw, &event); err != nil {
		return false
	}
	return matcher(event)
}

func isObject(v any) bool {
	_, ok := v.(map[string]any)
	return ok
}

func isArray(v any) bool {
	_, ok := v.([]any)
	return ok
}

func nestedObject(root map[string]any, keys ...string) map[string]any {
	current := root
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func toString(v any, keys ...string) string {
	obj, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	current := obj
	for i, key := range keys {
		value, exists := current[key]
		if !exists {
			return ""
		}
		if i == len(keys)-1 {
			s, _ := value.(string)
			return s
		}
		next, ok := value.(map[string]any)
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

func hasString(root map[string]any, key string) bool {
	_, ok := root[key].(string)
	return ok
}

func hasCodeCommitRecord(envelope map[string]any) bool {
	records, ok := envelope["Records"].([]any)
	if !ok || len(records) == 0 {
		return false
	}
	first, ok := records[0].(map[string]any)
	if !ok {
		return false
	}
	return isObject(first["codecommit"])
}

func isCloudFrontRequestEvent(envelope map[string]any) bool {
	records, ok := envelope["Records"].([]any)
	if !ok || len(records) == 0 {
		return false
	}
	first, ok := records[0].(map[string]any)
	if !ok {
		return false
	}
	cf, ok := first["cf"].(map[string]any)
	if !ok {
		return false
	}
	_, hasReq := cf["request"].(map[string]any)
	_, hasResp := cf["response"].(map[string]any)
	return hasReq && !hasResp
}

func isCloudFrontResponseEvent(envelope map[string]any) bool {
	records, ok := envelope["Records"].([]any)
	if !ok || len(records) == 0 {
		return false
	}
	first, ok := records[0].(map[string]any)
	if !ok {
		return false
	}
	cf, ok := first["cf"].(map[string]any)
	if !ok {
		return false
	}
	_, hasResp := cf["response"].(map[string]any)
	return hasResp
}

func isAPIGatewayAuthorizerEvent(envelope map[string]any) bool {
	eventType, ok := envelope["type"].(string)
	if !ok || (eventType != "TOKEN" && eventType != "REQUEST") {
		return false
	}
	return hasString(envelope, "methodArn") || hasString(envelope, "routeArn")
}

func isCloudFormationCustomResourceEvent(envelope map[string]any) bool {
	required := []string{
		"RequestType",
		"ResponseURL",
		"StackId",
		"RequestId",
		"LogicalResourceId",
		"ResourceType",
	}
	for _, key := range required {
		if !hasString(envelope, key) {
			return false
		}
	}
	return true
}
