package events

import (
	"encoding/json"
	"testing"
)

func parse(t *testing.T, raw string) (json.RawMessage, map[string]any) {
	t.Helper()
	r := json.RawMessage(raw)
	env, err := ParseEnvelope(r)
	if err != nil {
		t.Fatalf("invalid test JSON: %v", err)
	}
	return r, env
}

func TestIsHTTPV2(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{
			"version": "2.0",
			"requestContext": { "http": { "method": "GET", "path": "/" } }
		}`)
		event, ok := IsHTTPV2(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
		if event.Version != "2.0" {
			t.Fatalf("expected version 2.0, got %q", event.Version)
		}
	})
	t.Run("reject v1", func(t *testing.T) {
		raw, env := parse(t, `{"httpMethod": "GET", "path": "/"}`)
		_, ok := IsHTTPV2(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsLambdaFunctionURL(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{
			"version": "2.0",
			"requestContext": {
				"http": { "method": "GET", "path": "/" },
				"domainName": "abc.lambda-url.us-east-1.on.aws"
			}
		}`)
		_, ok := IsLambdaFunctionURL(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject regular v2", func(t *testing.T) {
		raw, env := parse(t, `{
			"version": "2.0",
			"requestContext": {
				"http": { "method": "GET", "path": "/" },
				"domainName": "api.example.com"
			}
		}`)
		_, ok := IsLambdaFunctionURL(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsALB(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"requestContext": {"elb": {}}}`)
		_, ok := IsALB(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"requestContext": {}}`)
		_, ok := IsALB(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsHTTPV1(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"httpMethod": "GET", "path": "/"}`)
		event, ok := IsHTTPV1(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
		if event.HTTPMethod != "GET" {
			t.Fatalf("expected GET, got %q", event.HTTPMethod)
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"foo": "bar"}`)
		_, ok := IsHTTPV1(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsSQS(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:sqs", "messageId": "abc"}]}`)
		event, ok := IsSQS(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
		if len(event.Records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(event.Records))
		}
	})
	t.Run("reject sns", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"EventSource": "aws:sns"}]}`)
		_, ok := IsSQS(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsSNS(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"EventSource": "aws:sns"}]}`)
		_, ok := IsSNS(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject sqs", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:sqs"}]}`)
		_, ok := IsSNS(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsS3(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:s3", "eventName": "ObjectCreated:Put"}]}`)
		_, ok := IsS3(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:sqs"}]}`)
		_, ok := IsS3(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsSES(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:ses", "ses": {}}]}`)
		_, ok := IsSES(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:s3"}]}`)
		_, ok := IsSES(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsDynamoDBStream(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:dynamodb", "eventName": "INSERT"}]}`)
		_, ok := IsDynamoDBStream(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:kinesis"}]}`)
		_, ok := IsDynamoDBStream(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsKinesisStream(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:kinesis", "eventName": "aws:kinesis:record"}]}`)
		_, ok := IsKinesisStream(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:dynamodb"}]}`)
		_, ok := IsKinesisStream(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCloudWatchLogs(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"awslogs": {"data": "H4sIAAAAAAAA..."}}`)
		_, ok := IsCloudWatchLogs(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"foo": "bar"}`)
		_, ok := IsCloudWatchLogs(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsScheduledEvent(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.events", "detail-type": "Scheduled Event"}`)
		event, ok := IsScheduledEvent(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
		if event.Source != "aws.events" {
			t.Fatalf("expected source aws.events, got %q", event.Source)
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"source": "custom.app", "detail-type": "User Created"}`)
		_, ok := IsScheduledEvent(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCodeBuildState(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.codebuild", "detail-type": "CodeBuild Build State Change"}`)
		_, ok := IsCodeBuildState(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.codebuild", "detail-type": "Other"}`)
		_, ok := IsCodeBuildState(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCodeCommit(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"codecommit": {"references": []}}]}`)
		_, ok := IsCodeCommit(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:sqs"}]}`)
		_, ok := IsCodeCommit(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCloudWatchAlarm(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"alarmArn": "arn:aws:cloudwatch:...", "alarmData": {}}`)
		result, ok := IsCloudWatchAlarm(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
		if result == nil {
			t.Fatal("expected non-nil raw message")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"alarmArn": "arn:aws:cloudwatch:..."}`)
		_, ok := IsCloudWatchAlarm(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsConnectContactFlow(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Name": "ContactFlowEvent", "Details": {"ContactData": {}}}`)
		_, ok := IsConnectContactFlow(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"Name": "Other"}`)
		_, ok := IsConnectContactFlow(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCloudFrontRequest(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"cf": {"request": {}}}]}`)
		_, ok := IsCloudFrontRequest(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject response event", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"cf": {"request": {}, "response": {}}}]}`)
		_, ok := IsCloudFrontRequest(raw, env)
		if ok {
			t.Fatal("expected no match for response event")
		}
	})
}

func TestIsCloudFrontResponse(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"cf": {"response": {}}}]}`)
		_, ok := IsCloudFrontResponse(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject no cf", func(t *testing.T) {
		raw, env := parse(t, `{"Records": [{"eventSource": "aws:sqs"}]}`)
		_, ok := IsCloudFrontResponse(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsGuardDuty(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.guardduty", "detail-type": "GuardDuty Finding"}`)
		_, ok := IsGuardDuty(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.events", "detail-type": "Scheduled Event"}`)
		_, ok := IsGuardDuty(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsS3Batch(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"invocationSchemaVersion": "1.0", "tasks": []}`)
		_, ok := IsS3Batch(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"invocationSchemaVersion": "1.0"}`)
		_, ok := IsS3Batch(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsS3Notification(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.s3", "detail-type": "Object Created"}`)
		_, ok := IsS3Notification(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"source": "aws.ec2", "detail-type": "Instance Terminated"}`)
		_, ok := IsS3Notification(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsLex(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"messageVersion": "1.0", "currentIntent": {}, "bot": {}}`)
		_, ok := IsLex(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject no intent", func(t *testing.T) {
		raw, env := parse(t, `{"messageVersion": "1.0", "bot": {}}`)
		_, ok := IsLex(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsLexV2(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"messageVersion": "1.0", "sessionId": "sid", "bot": {}}`)
		_, ok := IsLexV2(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject no session", func(t *testing.T) {
		raw, env := parse(t, `{"messageVersion": "1.0", "bot": {}}`)
		_, ok := IsLexV2(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsAppSyncResolver(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"info": {"fieldName": "listPosts"}, "request": {}}`)
		_, ok := IsAppSyncResolver(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject no info", func(t *testing.T) {
		raw, env := parse(t, `{"request": {}}`)
		_, ok := IsAppSyncResolver(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsAmplifyResolver(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"typeName": "Query", "fieldName": "listPosts", "request": {}}`)
		_, ok := IsAmplifyResolver(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject missing field", func(t *testing.T) {
		raw, env := parse(t, `{"typeName": "Query", "request": {}}`)
		_, ok := IsAmplifyResolver(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsAPIGatewayAuthorizer(t *testing.T) {
	t.Run("match token", func(t *testing.T) {
		raw, env := parse(t, `{"type": "TOKEN", "methodArn": "arn:aws:execute-api:..."}`)
		_, ok := IsAPIGatewayAuthorizer(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject wrong type", func(t *testing.T) {
		raw, env := parse(t, `{"type": "INVALID", "methodArn": "arn:aws:execute-api:..."}`)
		_, ok := IsAPIGatewayAuthorizer(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsAutoScalingScaleIn(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"AutoScalingGroupARN": "arn:aws:autoscaling:...", "CapacityToTerminate": []}`)
		_, ok := IsAutoScalingScaleIn(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"AutoScalingGroupARN": "arn:aws:autoscaling:..."}`)
		_, ok := IsAutoScalingScaleIn(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCloudFormationCustomResource(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{
			"RequestType": "Create", "ResponseURL": "https://example.com",
			"StackId": "stack", "RequestId": "req",
			"LogicalResourceId": "MyResource", "ResourceType": "Custom::MyType"
		}`)
		_, ok := IsCloudFormationCustomResource(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject missing field", func(t *testing.T) {
		raw, env := parse(t, `{"RequestType": "Create", "ResponseURL": "https://example.com"}`)
		_, ok := IsCloudFormationCustomResource(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsIoTCustomAuthorizer(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"protocolData": {}, "connectionMetadata": {}}`)
		_, ok := IsIoTCustomAuthorizer(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"protocolData": {}}`)
		_, ok := IsIoTCustomAuthorizer(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsEventBridge(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"source": "custom.app", "detail-type": "User Created"}`)
		_, ok := IsEventBridge(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject empty source", func(t *testing.T) {
		raw, env := parse(t, `{"source": "", "detail-type": "User Created"}`)
		_, ok := IsEventBridge(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsCodePipeline(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"CodePipeline.job": {"id": "job-123"}}`)
		_, ok := IsCodePipeline(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"foo": "bar"}`)
		_, ok := IsCodePipeline(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsFirehoseTransformation(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"invocationId": "inv-1", "records": [{"recordId": "r1", "data": "Zm9v"}]}`)
		_, ok := IsFirehoseTransformation(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject no records", func(t *testing.T) {
		raw, env := parse(t, `{"invocationId": "inv-1"}`)
		_, ok := IsFirehoseTransformation(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsSecretsManagerRotation(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"Step": "createSecret", "SecretId": "arn:aws:secretsmanager:...", "ClientRequestToken": "token-1"}`)
		_, ok := IsSecretsManagerRotation(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject missing token", func(t *testing.T) {
		raw, env := parse(t, `{"Step": "createSecret", "SecretId": "arn:aws:secretsmanager:..."}`)
		_, ok := IsSecretsManagerRotation(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsTransferFamilyAuth(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{"username": "u1", "serverId": "s-123", "sourceIp": "127.0.0.1"}`)
		_, ok := IsTransferFamilyAuth(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject", func(t *testing.T) {
		raw, env := parse(t, `{"username": "u1"}`)
		_, ok := IsTransferFamilyAuth(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsMSK(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{
			"eventSource": "aws:kafka",
			"records": {"topic-0": [{"topic": "topic", "partition": 0}]}
		}`)
		_, ok := IsMSK(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject self-managed", func(t *testing.T) {
		raw, env := parse(t, `{
			"eventSource": "SelfManagedKafka",
			"records": {"topic-0": [{"topic": "topic"}]}
		}`)
		_, ok := IsMSK(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}

func TestIsSelfManagedKafka(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		raw, env := parse(t, `{
			"eventSource": "SelfManagedKafka",
			"records": {"topic-0": [{"topic": "topic"}]}
		}`)
		_, ok := IsSelfManagedKafka(raw, env)
		if !ok {
			t.Fatal("expected match")
		}
	})
	t.Run("reject msk", func(t *testing.T) {
		raw, env := parse(t, `{
			"eventSource": "aws:kafka",
			"records": {"topic-0": [{"topic": "topic"}]}
		}`)
		_, ok := IsSelfManagedKafka(raw, env)
		if ok {
			t.Fatal("expected no match")
		}
	})
}
