package lambdaapp

import "testing"

func TestDetectEventType(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected EventType
		wantErr  bool
	}{
		{
			name: "http v2",
			raw: `{
				"version": "2.0",
				"requestContext": {
					"http": { "method": "GET", "path": "/" }
				}
			}`,
			expected: EventTypeHTTPV2,
		},
		{
			name: "lambda function url",
			raw: `{
				"version": "2.0",
				"requestContext": {
					"http": { "method": "GET", "path": "/" },
					"domainName": "abc.lambda-url.us-east-1.on.aws"
				}
			}`,
			expected: EventTypeLambdaFunctionURL,
		},
		{
			name: "http v1",
			raw: `{
				"httpMethod": "GET",
				"path": "/"
			}`,
			expected: EventTypeHTTPV1,
		},
		{
			name: "sqs",
			raw: `{
				"Records": [
					{ "eventSource": "aws:sqs", "messageId": "abc" }
				]
			}`,
			expected: EventTypeSQS,
		},
		{
			name: "sns with EventSource",
			raw: `{
				"Records": [
					{ "EventSource": "aws:sns" }
				]
			}`,
			expected: EventTypeSNS,
		},
		{
			name: "alb",
			raw: `{
				"requestContext": {
					"elb": {}
				}
			}`,
			expected: EventTypeALB,
		},
		{
			name: "scheduled event",
			raw: `{
				"source": "aws.events",
				"detail-type": "Scheduled Event"
			}`,
			expected: EventTypeScheduled,
		},
		{
			name: "codebuild state event",
			raw: `{
				"source": "aws.codebuild",
				"detail-type": "CodeBuild Build State Change"
			}`,
			expected: EventTypeCodeBuildState,
		},
		{
			name: "codecommit event",
			raw: `{
				"Records": [
					{ "codecommit": { "references": [] } }
				]
			}`,
			expected: EventTypeCodeCommit,
		},
		{
			name: "eventbridge",
			raw: `{
				"source": "custom.app",
				"detail-type": "User Created"
			}`,
			expected: EventTypeEventBridge,
		},
		{
			name: "cloudwatch logs",
			raw: `{
				"awslogs": {
					"data": "H4sIAAAAAAAA..."
				}
			}`,
			expected: EventTypeCloudWatchLogs,
		},
		{
			name: "cloudwatch alarm",
			raw: `{
				"alarmArn": "arn:aws:cloudwatch:...",
				"alarmData": {}
			}`,
			expected: EventTypeCloudWatchAlarm,
		},
		{
			name: "connect contact flow",
			raw: `{
				"Name": "ContactFlowEvent",
				"Details": { "ContactData": {} }
			}`,
			expected: EventTypeConnectContactFlow,
		},
		{
			name: "cloudfront request",
			raw: `{
				"Records": [
					{ "cf": { "request": {} } }
				]
			}`,
			expected: EventTypeCloudFrontRequest,
		},
		{
			name: "cloudfront response",
			raw: `{
				"Records": [
					{ "cf": { "response": {} } }
				]
			}`,
			expected: EventTypeCloudFrontResponse,
		},
		{
			name: "guardduty notification",
			raw: `{
				"source": "aws.guardduty",
				"detail-type": "GuardDuty Finding"
			}`,
			expected: EventTypeGuardDutyNotification,
		},
		{
			name: "s3 batch",
			raw: `{
				"invocationSchemaVersion": "1.0",
				"tasks": []
			}`,
			expected: EventTypeS3Batch,
		},
		{
			name: "s3 notification eventbridge",
			raw: `{
				"source": "aws.s3",
				"detail-type": "Object Created"
			}`,
			expected: EventTypeS3Notification,
		},
		{
			name: "lex v1",
			raw: `{
				"messageVersion": "1.0",
				"currentIntent": {},
				"bot": {}
			}`,
			expected: EventTypeLex,
		},
		{
			name: "lex v2",
			raw: `{
				"messageVersion": "1.0",
				"sessionId": "sid",
				"bot": {}
			}`,
			expected: EventTypeLexV2,
		},
		{
			name: "appsync resolver",
			raw: `{
				"info": { "fieldName": "listPosts" },
				"request": {}
			}`,
			expected: EventTypeAppSyncResolver,
		},
		{
			name: "amplify resolver",
			raw: `{
				"typeName": "Query",
				"fieldName": "listPosts",
				"request": {}
			}`,
			expected: EventTypeAmplifyResolver,
		},
		{
			name: "api gateway authorizer",
			raw: `{
				"type": "TOKEN",
				"methodArn": "arn:aws:execute-api:..."
			}`,
			expected: EventTypeAPIGatewayAuthorizer,
		},
		{
			name: "autoscaling scale-in",
			raw: `{
				"AutoScalingGroupARN": "arn:aws:autoscaling:...",
				"CapacityToTerminate": []
			}`,
			expected: EventTypeAutoScalingScaleIn,
		},
		{
			name: "cloudformation custom resource",
			raw: `{
				"RequestType": "Create",
				"ResponseURL": "https://example.com",
				"StackId": "stack",
				"RequestId": "req",
				"LogicalResourceId": "MyResource",
				"ResourceType": "Custom::MyType"
			}`,
			expected: EventTypeCloudFormationCustom,
		},
		{
			name: "iot scalar event",
			raw:  `"temperature:42"`,
			expected: EventTypeIoT,
		},
		{
			name: "iot custom authorizer",
			raw: `{
				"protocolData": {},
				"connectionMetadata": {}
			}`,
			expected: EventTypeIoTCustomAuthorizer,
		},
		{
			name: "transfer family authorizer",
			raw: `{
				"username": "u1",
				"serverId": "s-123",
				"sourceIp": "127.0.0.1"
			}`,
			expected: EventTypeTransferFamilyAuth,
		},
		{
			name: "codepipeline",
			raw: `{
				"CodePipeline.job": {
					"id": "job-123"
				}
			}`,
			expected: EventTypeCodePipeline,
		},
		{
			name: "firehose transformation",
			raw: `{
				"invocationId": "inv-1",
				"records": [
					{ "recordId": "r1", "data": "Zm9v" }
				]
			}`,
			expected: EventTypeFirehoseTransformation,
		},
		{
			name: "secrets manager rotation",
			raw: `{
				"Step": "createSecret",
				"SecretId": "arn:aws:secretsmanager:...",
				"ClientRequestToken": "token-1"
			}`,
			expected: EventTypeSecretsManagerRotation,
		},
		{
			name: "msk",
			raw: `{
				"eventSource": "aws:kafka",
				"records": {
					"topic-0": [
						{ "topic": "topic", "partition": 0, "offset": 1, "timestamp": 1, "timestampType": "CREATE_TIME" }
					]
				}
			}`,
			expected: EventTypeMSK,
		},
		{
			name: "self managed kafka",
			raw: `{
				"eventSource": "SelfManagedKafka",
				"records": {
					"topic-0": [
						{ "topic": "topic", "partition": 0, "offset": 1, "timestamp": 1, "timestampType": "CREATE_TIME" }
					]
				}
			}`,
			expected: EventTypeSelfManagedKafka,
		},
		{
			name: "s3 classic records event",
			raw: `{
				"Records": [
					{ "eventSource": "aws:s3", "eventName": "ObjectCreated:Put" }
				]
			}`,
			expected: EventTypeS3,
		},
		{
			name: "ses records event",
			raw: `{
				"Records": [
					{ "eventSource": "aws:ses", "ses": {} }
				]
			}`,
			expected: EventTypeSES,
		},
		{
			name: "dynamodb stream",
			raw: `{
				"Records": [
					{ "eventSource": "aws:dynamodb", "eventName": "INSERT" }
				]
			}`,
			expected: EventTypeDynamoDBStream,
		},
		{
			name: "kinesis stream",
			raw: `{
				"Records": [
					{ "eventSource": "aws:kinesis", "eventName": "aws:kinesis:record" }
				]
			}`,
			expected: EventTypeKinesisStream,
		},
		{
			name: "iot numeric event",
			raw:  `42`,
			expected: EventTypeIoT,
		},
		{
			name: "non-object non-scalar event",
			raw:  `[]`,
			expected: EventTypeUnknown,
		},
		{
			name: "unknown event",
			raw: `{
				"foo": "bar"
			}`,
			expected: EventTypeUnknown,
		},
		{
			name:    "invalid json",
			raw:     `{`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := detectEventType([]byte(tc.raw))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (type=%q)", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got != tc.expected {
				t.Fatalf("expected type=%q, got %q", tc.expected, got)
			}
		})
	}
}
