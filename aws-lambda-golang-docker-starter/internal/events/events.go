package events

import (
	"encoding/json"
	"strings"

	awsevents "github.com/aws/aws-lambda-go/events"
)

func ParseEnvelope(raw json.RawMessage) (map[string]any, error) {
	var envelope map[string]any
	err := json.Unmarshal(raw, &envelope)
	return envelope, err
}

func unmarshalAs[T any](raw json.RawMessage, match bool) (T, bool) {
	var event T
	if !match {
		return event, false
	}
	if err := json.Unmarshal(raw, &event); err != nil {
		return event, false
	}
	return event, true
}

func matchRaw(raw json.RawMessage, match bool) (json.RawMessage, bool) {
	if !match {
		return nil, false
	}
	return raw, true
}

func IsHTTPV2(raw json.RawMessage, envelope map[string]any) (awsevents.APIGatewayV2HTTPRequest, bool) {
	version, _ := envelope["version"].(string)
	if version != "2.0" {
		return awsevents.APIGatewayV2HTTPRequest{}, false
	}
	httpCtx := nestedObject(envelope, "requestContext", "http")
	if httpCtx == nil {
		return awsevents.APIGatewayV2HTTPRequest{}, false
	}
	method, _ := httpCtx["method"].(string)
	return unmarshalAs[awsevents.APIGatewayV2HTTPRequest](raw, strings.TrimSpace(method) != "")
}

func IsLambdaFunctionURL(raw json.RawMessage, envelope map[string]any) (awsevents.APIGatewayV2HTTPRequest, bool) {
	event, ok := IsHTTPV2(raw, envelope)
	if !ok {
		return event, false
	}
	rc := nestedObject(envelope, "requestContext")
	if rc == nil {
		return awsevents.APIGatewayV2HTTPRequest{}, false
	}
	domainName, _ := rc["domainName"].(string)
	if !strings.Contains(domainName, "lambda-url") {
		return awsevents.APIGatewayV2HTTPRequest{}, false
	}
	return event, true
}

func IsALB(raw json.RawMessage, envelope map[string]any) (awsevents.ALBTargetGroupRequest, bool) {
	return unmarshalAs[awsevents.ALBTargetGroupRequest](raw, nestedObject(envelope, "requestContext", "elb") != nil)
}

func IsHTTPV1(raw json.RawMessage, envelope map[string]any) (awsevents.APIGatewayProxyRequest, bool) {
	method, _ := envelope["httpMethod"].(string)
	return unmarshalAs[awsevents.APIGatewayProxyRequest](raw, strings.TrimSpace(method) != "")
}

func IsSQS(raw json.RawMessage, envelope map[string]any) (awsevents.SQSEvent, bool) {
	return unmarshalAs[awsevents.SQSEvent](raw, firstRecordField(envelope, "eventSource") == "aws:sqs")
}

func IsSNS(raw json.RawMessage, envelope map[string]any) (awsevents.SNSEvent, bool) {
	return unmarshalAs[awsevents.SNSEvent](raw, firstRecordField(envelope, "EventSource") == "aws:sns")
}

func IsS3(raw json.RawMessage, envelope map[string]any) (awsevents.S3Event, bool) {
	return unmarshalAs[awsevents.S3Event](raw, firstRecordField(envelope, "eventSource") == "aws:s3")
}

func IsSES(raw json.RawMessage, envelope map[string]any) (awsevents.SimpleEmailEvent, bool) {
	return unmarshalAs[awsevents.SimpleEmailEvent](raw, firstRecordField(envelope, "eventSource") == "aws:ses")
}

func IsDynamoDBStream(raw json.RawMessage, envelope map[string]any) (awsevents.DynamoDBEvent, bool) {
	return unmarshalAs[awsevents.DynamoDBEvent](raw, firstRecordField(envelope, "eventSource") == "aws:dynamodb")
}

func IsKinesisStream(raw json.RawMessage, envelope map[string]any) (awsevents.KinesisEvent, bool) {
	return unmarshalAs[awsevents.KinesisEvent](raw, firstRecordField(envelope, "eventSource") == "aws:kinesis")
}

func IsCloudWatchLogs(raw json.RawMessage, envelope map[string]any) (awsevents.CloudwatchLogsEvent, bool) {
	logs := nestedObject(envelope, "awslogs")
	if logs == nil {
		return awsevents.CloudwatchLogsEvent{}, false
	}
	data, _ := logs["data"].(string)
	return unmarshalAs[awsevents.CloudwatchLogsEvent](raw, strings.TrimSpace(data) != "")
}

func IsScheduledEvent(raw json.RawMessage, envelope map[string]any) (awsevents.CloudWatchEvent, bool) {
	source, _ := envelope["source"].(string)
	detailType, _ := envelope["detail-type"].(string)
	return unmarshalAs[awsevents.CloudWatchEvent](raw, source == "aws.events" && detailType == "Scheduled Event")
}

func IsCodeBuildState(raw json.RawMessage, envelope map[string]any) (awsevents.CodeBuildEvent, bool) {
	source, _ := envelope["source"].(string)
	detailType, _ := envelope["detail-type"].(string)
	return unmarshalAs[awsevents.CodeBuildEvent](raw, source == "aws.codebuild" && detailType == "CodeBuild Build State Change")
}

func IsCodeCommit(raw json.RawMessage, envelope map[string]any) (awsevents.CodeCommitEvent, bool) {
	first := firstRecord(envelope)
	if first == nil {
		return awsevents.CodeCommitEvent{}, false
	}
	_, ok := first["codecommit"].(map[string]any)
	return unmarshalAs[awsevents.CodeCommitEvent](raw, ok)
}

func IsCloudWatchAlarm(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	_, hasArn := envelope["alarmArn"].(string)
	_, hasData := envelope["alarmData"].(map[string]any)
	return matchRaw(raw, hasArn && hasData)
}

func IsConnectContactFlow(raw json.RawMessage, envelope map[string]any) (awsevents.ConnectEvent, bool) {
	name, _ := envelope["Name"].(string)
	return unmarshalAs[awsevents.ConnectEvent](raw, name == "ContactFlowEvent" && nestedObject(envelope, "Details", "ContactData") != nil)
}

func IsCloudFrontRequest(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	cf := firstRecordCF(envelope)
	if cf == nil {
		return nil, false
	}
	_, hasReq := cf["request"].(map[string]any)
	_, hasResp := cf["response"].(map[string]any)
	return matchRaw(raw, hasReq && !hasResp)
}

func IsCloudFrontResponse(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	cf := firstRecordCF(envelope)
	if cf == nil {
		return nil, false
	}
	_, hasResp := cf["response"].(map[string]any)
	return matchRaw(raw, hasResp)
}

func IsGuardDuty(raw json.RawMessage, envelope map[string]any) (awsevents.CloudWatchEvent, bool) {
	source, _ := envelope["source"].(string)
	return unmarshalAs[awsevents.CloudWatchEvent](raw, source == "aws.guardduty")
}

func IsS3Batch(raw json.RawMessage, envelope map[string]any) (awsevents.S3BatchJobEvent, bool) {
	_, hasSchema := envelope["invocationSchemaVersion"].(string)
	_, hasTasks := envelope["tasks"].([]any)
	return unmarshalAs[awsevents.S3BatchJobEvent](raw, hasSchema && hasTasks)
}

func IsS3Notification(raw json.RawMessage, envelope map[string]any) (awsevents.CloudWatchEvent, bool) {
	source, _ := envelope["source"].(string)
	detailType, _ := envelope["detail-type"].(string)
	return unmarshalAs[awsevents.CloudWatchEvent](raw, source == "aws.s3" && strings.TrimSpace(detailType) != "")
}

func IsLex(raw json.RawMessage, envelope map[string]any) (awsevents.LexEvent, bool) {
	_, hasMsgVer := envelope["messageVersion"].(string)
	_, hasIntent := envelope["currentIntent"].(map[string]any)
	_, hasBot := envelope["bot"].(map[string]any)
	return unmarshalAs[awsevents.LexEvent](raw, hasMsgVer && hasIntent && hasBot)
}

func IsLexV2(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	_, hasMsgVer := envelope["messageVersion"].(string)
	_, hasBot := envelope["bot"].(map[string]any)
	_, hasSession := envelope["sessionId"].(string)
	return matchRaw(raw, hasMsgVer && hasBot && hasSession)
}

func IsAppSyncResolver(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	info := nestedObject(envelope, "info")
	if info == nil {
		return nil, false
	}
	_, hasField := info["fieldName"].(string)
	_, hasReq := envelope["request"].(map[string]any)
	return matchRaw(raw, hasField && hasReq)
}

func IsAmplifyResolver(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	_, hasType := envelope["typeName"].(string)
	_, hasField := envelope["fieldName"].(string)
	_, hasReq := envelope["request"].(map[string]any)
	return matchRaw(raw, hasType && hasField && hasReq)
}

func IsAPIGatewayAuthorizer(raw json.RawMessage, envelope map[string]any) (awsevents.APIGatewayCustomAuthorizerRequest, bool) {
	eventType, ok := envelope["type"].(string)
	if !ok || (eventType != "TOKEN" && eventType != "REQUEST") {
		return awsevents.APIGatewayCustomAuthorizerRequest{}, false
	}
	_, hasMethod := envelope["methodArn"].(string)
	_, hasRoute := envelope["routeArn"].(string)
	return unmarshalAs[awsevents.APIGatewayCustomAuthorizerRequest](raw, hasMethod || hasRoute)
}

func IsAutoScalingScaleIn(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	_, hasARN := envelope["AutoScalingGroupARN"].(string)
	_, hasCap := envelope["CapacityToTerminate"].([]any)
	return matchRaw(raw, hasARN && hasCap)
}

func IsCloudFormationCustomResource(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	for _, key := range []string{"RequestType", "ResponseURL", "StackId", "RequestId", "LogicalResourceId", "ResourceType"} {
		if _, ok := envelope[key].(string); !ok {
			return nil, false
		}
	}
	return raw, true
}

func IsIoTCustomAuthorizer(raw json.RawMessage, envelope map[string]any) (awsevents.IoTCoreCustomAuthorizerRequest, bool) {
	_, hasProto := envelope["protocolData"].(map[string]any)
	_, hasConn := envelope["connectionMetadata"].(map[string]any)
	return unmarshalAs[awsevents.IoTCoreCustomAuthorizerRequest](raw, hasProto && hasConn)
}

func IsEventBridge(raw json.RawMessage, envelope map[string]any) (awsevents.CloudWatchEvent, bool) {
	source, _ := envelope["source"].(string)
	detailType, _ := envelope["detail-type"].(string)
	return unmarshalAs[awsevents.CloudWatchEvent](raw, strings.TrimSpace(source) != "" && strings.TrimSpace(detailType) != "")
}

func IsCodePipeline(raw json.RawMessage, envelope map[string]any) (awsevents.CodePipelineEvent, bool) {
	job := nestedObject(envelope, "CodePipeline.job")
	if job == nil {
		return awsevents.CodePipelineEvent{}, false
	}
	id, _ := job["id"].(string)
	return unmarshalAs[awsevents.CodePipelineEvent](raw, strings.TrimSpace(id) != "")
}

func IsFirehoseTransformation(raw json.RawMessage, envelope map[string]any) (awsevents.KinesisFirehoseEvent, bool) {
	invocationID, _ := envelope["invocationId"].(string)
	records, _ := envelope["records"].([]any)
	return unmarshalAs[awsevents.KinesisFirehoseEvent](raw, strings.TrimSpace(invocationID) != "" && len(records) > 0)
}

func IsSecretsManagerRotation(raw json.RawMessage, envelope map[string]any) (awsevents.SecretsManagerSecretRotationEvent, bool) {
	step, _ := envelope["Step"].(string)
	secretID, _ := envelope["SecretId"].(string)
	token, _ := envelope["ClientRequestToken"].(string)
	return unmarshalAs[awsevents.SecretsManagerSecretRotationEvent](raw,
		strings.TrimSpace(step) != "" && strings.TrimSpace(secretID) != "" && strings.TrimSpace(token) != "")
}

func IsTransferFamilyAuth(raw json.RawMessage, envelope map[string]any) (json.RawMessage, bool) {
	_, hasUser := envelope["username"].(string)
	_, hasServer := envelope["serverId"].(string)
	_, hasIP := envelope["sourceIp"].(string)
	return matchRaw(raw, hasUser && hasServer && hasIP)
}

func IsMSK(raw json.RawMessage, envelope map[string]any) (awsevents.KafkaEvent, bool) {
	source, _ := envelope["eventSource"].(string)
	records, _ := envelope["records"].(map[string]any)
	return unmarshalAs[awsevents.KafkaEvent](raw, source == "aws:kafka" && len(records) > 0)
}

func IsSelfManagedKafka(raw json.RawMessage, envelope map[string]any) (awsevents.KafkaEvent, bool) {
	source, _ := envelope["eventSource"].(string)
	records, _ := envelope["records"].(map[string]any)
	return unmarshalAs[awsevents.KafkaEvent](raw, source == "SelfManagedKafka" && len(records) > 0)
}

// --- helpers ---

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

func firstRecord(envelope map[string]any) map[string]any {
	records, ok := envelope["Records"].([]any)
	if !ok || len(records) == 0 {
		return nil
	}
	first, _ := records[0].(map[string]any)
	return first
}

func firstRecordField(envelope map[string]any, key string) string {
	rec := firstRecord(envelope)
	if rec == nil {
		return ""
	}
	val, _ := rec[key].(string)
	return val
}

func firstRecordCF(envelope map[string]any) map[string]any {
	rec := firstRecord(envelope)
	if rec == nil {
		return nil
	}
	cf, _ := rec["cf"].(map[string]any)
	return cf
}
