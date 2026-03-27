"""Tests for app.events – mirrors the Go events_test.go structure."""
from __future__ import annotations

import pytest

from app.events import (
    create_eventbridge_matcher,
    is_alb,
    is_amplify_resolver,
    is_api_gateway_authorizer,
    is_appsync_resolver,
    is_autoscaling_scale_in,
    is_cloudformation_custom_resource,
    is_cloudfront_request,
    is_cloudfront_response,
    is_cloudwatch_alarm,
    is_cloudwatch_logs,
    is_codebuild_state,
    is_codecommit,
    is_codepipeline,
    is_connect_contact_flow,
    is_dynamodb_stream,
    is_eventbridge,
    is_firehose_transformation,
    is_guardduty,
    is_http_event,
    is_http_v1,
    is_http_v2,
    is_iot_custom_authorizer,
    is_kinesis_stream,
    is_lambda_function_url,
    is_lex,
    is_lex_v2,
    is_msk,
    is_s3,
    is_s3_batch,
    is_s3_notification,
    is_scheduled_event,
    is_secrets_manager_rotation,
    is_self_managed_kafka,
    is_ses,
    is_sns,
    is_sqs,
    is_transfer_family_auth,
)


# ---------- HTTP-family ----------

class TestIsHttpV2:
    def test_match(self) -> None:
        event = {
            "version": "2.0",
            "requestContext": {"http": {"method": "GET", "path": "/"}},
        }
        assert is_http_v2(event) is True

    def test_reject_v1(self) -> None:
        assert is_http_v2({"httpMethod": "GET", "path": "/"}) is False

    def test_reject_missing_method(self) -> None:
        event = {"version": "2.0", "requestContext": {"http": {}}}
        assert is_http_v2(event) is False


class TestIsLambdaFunctionUrl:
    def test_match(self) -> None:
        event = {
            "version": "2.0",
            "requestContext": {
                "http": {"method": "GET", "path": "/"},
                "domainName": "abc.lambda-url.us-east-1.on.aws",
            },
        }
        assert is_lambda_function_url(event) is True

    def test_reject_regular_v2(self) -> None:
        event = {
            "version": "2.0",
            "requestContext": {
                "http": {"method": "GET", "path": "/"},
                "domainName": "api.example.com",
            },
        }
        assert is_lambda_function_url(event) is False


class TestIsAlb:
    def test_match(self) -> None:
        assert is_alb({"requestContext": {"elb": {}}}) is True

    def test_reject(self) -> None:
        assert is_alb({"requestContext": {}}) is False


class TestIsHttpV1:
    def test_match(self) -> None:
        event = {"httpMethod": "GET", "path": "/"}
        assert is_http_v1(event) is True

    def test_reject(self) -> None:
        assert is_http_v1({"foo": "bar"}) is False


class TestIsHttpEvent:
    def test_v1(self) -> None:
        assert is_http_event({"httpMethod": "POST", "path": "/"}) is True

    def test_v2(self) -> None:
        event = {"version": "2.0", "requestContext": {"http": {"method": "GET"}}}
        assert is_http_event(event) is True

    def test_alb(self) -> None:
        assert is_http_event({"requestContext": {"elb": {}}}) is True

    def test_reject(self) -> None:
        assert is_http_event({"foo": "bar"}) is False


# ---------- Queue / stream / storage ----------

class TestIsSqs:
    def test_match(self) -> None:
        event = {"Records": [{"eventSource": "aws:sqs", "messageId": "abc"}]}
        assert is_sqs(event) is True

    def test_reject_sns(self) -> None:
        event = {"Records": [{"EventSource": "aws:sns"}]}
        assert is_sqs(event) is False


class TestIsSns:
    def test_match(self) -> None:
        assert is_sns({"Records": [{"EventSource": "aws:sns"}]}) is True

    def test_reject_sqs(self) -> None:
        assert is_sns({"Records": [{"eventSource": "aws:sqs"}]}) is False


class TestIsS3:
    def test_match(self) -> None:
        event = {"Records": [{"eventSource": "aws:s3", "eventName": "ObjectCreated:Put"}]}
        assert is_s3(event) is True

    def test_reject(self) -> None:
        assert is_s3({"Records": [{"eventSource": "aws:sqs"}]}) is False


class TestIsSes:
    def test_match(self) -> None:
        assert is_ses({"Records": [{"eventSource": "aws:ses", "ses": {}}]}) is True

    def test_reject(self) -> None:
        assert is_ses({"Records": [{"eventSource": "aws:s3"}]}) is False


class TestIsDynamodbStream:
    def test_match(self) -> None:
        event = {"Records": [{"eventSource": "aws:dynamodb", "eventName": "INSERT"}]}
        assert is_dynamodb_stream(event) is True

    def test_reject(self) -> None:
        assert is_dynamodb_stream({"Records": [{"eventSource": "aws:kinesis"}]}) is False


class TestIsKinesisStream:
    def test_match(self) -> None:
        event = {"Records": [{"eventSource": "aws:kinesis", "eventName": "aws:kinesis:record"}]}
        assert is_kinesis_stream(event) is True

    def test_reject(self) -> None:
        assert is_kinesis_stream({"Records": [{"eventSource": "aws:dynamodb"}]}) is False


class TestIsMsk:
    def test_match(self) -> None:
        event = {"eventSource": "aws:kafka", "records": {"topic-0": [{"topic": "topic", "partition": 0}]}}
        assert is_msk(event) is True

    def test_reject_self_managed(self) -> None:
        event = {"eventSource": "SelfManagedKafka", "records": {"topic-0": [{"topic": "topic"}]}}
        assert is_msk(event) is False


class TestIsSelfManagedKafka:
    def test_match(self) -> None:
        event = {"eventSource": "SelfManagedKafka", "records": {"topic-0": [{"topic": "topic"}]}}
        assert is_self_managed_kafka(event) is True

    def test_reject_msk(self) -> None:
        event = {"eventSource": "aws:kafka", "records": {"topic-0": [{"topic": "topic"}]}}
        assert is_self_managed_kafka(event) is False


# ---------- CloudWatch / EventBridge ----------

class TestIsCloudwatchLogs:
    def test_match(self) -> None:
        assert is_cloudwatch_logs({"awslogs": {"data": "H4sIAAAAAAAA..."}}) is True

    def test_reject(self) -> None:
        assert is_cloudwatch_logs({"foo": "bar"}) is False

    def test_reject_empty_data(self) -> None:
        assert is_cloudwatch_logs({"awslogs": {"data": "  "}}) is False


class TestIsScheduledEvent:
    def test_match(self) -> None:
        event = {"source": "aws.events", "detail-type": "Scheduled Event"}
        assert is_scheduled_event(event) is True

    def test_reject(self) -> None:
        event = {"source": "custom.app", "detail-type": "User Created"}
        assert is_scheduled_event(event) is False


class TestIsCodebuildState:
    def test_match(self) -> None:
        event = {"source": "aws.codebuild", "detail-type": "CodeBuild Build State Change"}
        assert is_codebuild_state(event) is True

    def test_reject(self) -> None:
        event = {"source": "aws.codebuild", "detail-type": "Other"}
        assert is_codebuild_state(event) is False


class TestIsCloudwatchAlarm:
    def test_match(self) -> None:
        event = {"alarmArn": "arn:aws:cloudwatch:...", "alarmData": {}}
        assert is_cloudwatch_alarm(event) is True

    def test_reject(self) -> None:
        assert is_cloudwatch_alarm({"alarmArn": "arn:aws:cloudwatch:..."}) is False


class TestIsEventbridge:
    def test_match(self) -> None:
        event = {"source": "custom.app", "detail-type": "User Created"}
        assert is_eventbridge(event) is True

    def test_reject_empty_source(self) -> None:
        assert is_eventbridge({"source": "", "detail-type": "User Created"}) is False

    def test_reject_missing_detail_type(self) -> None:
        assert is_eventbridge({"source": "custom.app"}) is False


class TestCreateEventbridgeMatcher:
    def test_match_with_source(self) -> None:
        matcher = create_eventbridge_matcher(source="custom.app")
        assert matcher({"source": "custom.app", "detail-type": "Foo"}) is True

    def test_reject_wrong_source(self) -> None:
        matcher = create_eventbridge_matcher(source="custom.app")
        assert matcher({"source": "other.app", "detail-type": "Foo"}) is False

    def test_match_with_source_and_detail_type(self) -> None:
        matcher = create_eventbridge_matcher(source="aws.s3", detail_type="Object Created")
        assert matcher({"source": "aws.s3", "detail-type": "Object Created"}) is True
        assert matcher({"source": "aws.s3", "detail-type": "Other"}) is False


# ---------- Code* ----------

class TestIsCodecommit:
    def test_match(self) -> None:
        event = {"Records": [{"codecommit": {"references": []}}]}
        assert is_codecommit(event) is True

    def test_reject(self) -> None:
        assert is_codecommit({"Records": [{"eventSource": "aws:sqs"}]}) is False


class TestIsCodepipeline:
    def test_match(self) -> None:
        assert is_codepipeline({"CodePipeline.job": {"id": "job-123"}}) is True

    def test_reject(self) -> None:
        assert is_codepipeline({"foo": "bar"}) is False


# ---------- CloudFront ----------

class TestIsCloudFrontRequest:
    def test_match(self) -> None:
        event = {"Records": [{"cf": {"request": {}}}]}
        assert is_cloudfront_request(event) is True

    def test_reject_response_event(self) -> None:
        event = {"Records": [{"cf": {"request": {}, "response": {}}}]}
        assert is_cloudfront_request(event) is False


class TestIsCloudFrontResponse:
    def test_match(self) -> None:
        event = {"Records": [{"cf": {"response": {}}}]}
        assert is_cloudfront_response(event) is True

    def test_reject_no_cf(self) -> None:
        assert is_cloudfront_response({"Records": [{"eventSource": "aws:sqs"}]}) is False


# ---------- Security ----------

class TestIsGuardduty:
    def test_match(self) -> None:
        event = {"source": "aws.guardduty", "detail-type": "GuardDuty Finding"}
        assert is_guardduty(event) is True

    def test_reject(self) -> None:
        assert is_guardduty({"source": "aws.events", "detail-type": "Scheduled Event"}) is False


class TestIsConnectContactFlow:
    def test_match(self) -> None:
        event = {"Name": "ContactFlowEvent", "Details": {"ContactData": {}}}
        assert is_connect_contact_flow(event) is True

    def test_reject(self) -> None:
        assert is_connect_contact_flow({"Name": "Other"}) is False


# ---------- AI / resolver / API auth ----------

class TestIsS3Batch:
    def test_match(self) -> None:
        event = {"invocationSchemaVersion": "1.0", "tasks": []}
        assert is_s3_batch(event) is True

    def test_reject(self) -> None:
        assert is_s3_batch({"invocationSchemaVersion": "1.0"}) is False


class TestIsS3Notification:
    def test_match(self) -> None:
        event = {"source": "aws.s3", "detail-type": "Object Created"}
        assert is_s3_notification(event) is True

    def test_reject(self) -> None:
        event = {"source": "aws.ec2", "detail-type": "Instance Terminated"}
        assert is_s3_notification(event) is False


class TestIsLex:
    def test_match(self) -> None:
        event = {"messageVersion": "1.0", "currentIntent": {}, "bot": {}}
        assert is_lex(event) is True

    def test_reject_no_intent(self) -> None:
        assert is_lex({"messageVersion": "1.0", "bot": {}}) is False


class TestIsLexV2:
    def test_match(self) -> None:
        event = {"messageVersion": "1.0", "sessionId": "sid", "bot": {}}
        assert is_lex_v2(event) is True

    def test_reject_no_session(self) -> None:
        assert is_lex_v2({"messageVersion": "1.0", "bot": {}}) is False


class TestIsAppsyncResolver:
    def test_match(self) -> None:
        event = {"info": {"fieldName": "listPosts"}, "request": {}}
        assert is_appsync_resolver(event) is True

    def test_reject_no_info(self) -> None:
        assert is_appsync_resolver({"request": {}}) is False


class TestIsAmplifyResolver:
    def test_match(self) -> None:
        event = {"typeName": "Query", "fieldName": "listPosts", "request": {}}
        assert is_amplify_resolver(event) is True

    def test_reject_missing_field(self) -> None:
        assert is_amplify_resolver({"typeName": "Query", "request": {}}) is False


class TestIsApiGatewayAuthorizer:
    def test_match_token(self) -> None:
        event = {"type": "TOKEN", "methodArn": "arn:aws:execute-api:..."}
        assert is_api_gateway_authorizer(event) is True

    def test_match_request(self) -> None:
        event = {"type": "REQUEST", "routeArn": "arn:aws:execute-api:..."}
        assert is_api_gateway_authorizer(event) is True

    def test_reject_wrong_type(self) -> None:
        event = {"type": "INVALID", "methodArn": "arn:aws:execute-api:..."}
        assert is_api_gateway_authorizer(event) is False


# ---------- Infrastructure / device ----------

class TestIsAutoscalingScaleIn:
    def test_match(self) -> None:
        event = {"AutoScalingGroupARN": "arn:aws:autoscaling:...", "CapacityToTerminate": []}
        assert is_autoscaling_scale_in(event) is True

    def test_reject(self) -> None:
        assert is_autoscaling_scale_in({"AutoScalingGroupARN": "arn:aws:autoscaling:..."}) is False


class TestIsCloudFormationCustomResource:
    def test_match(self) -> None:
        event = {
            "RequestType": "Create",
            "ResponseURL": "https://example.com",
            "StackId": "stack",
            "RequestId": "req",
            "LogicalResourceId": "MyResource",
            "ResourceType": "Custom::MyType",
        }
        assert is_cloudformation_custom_resource(event) is True

    def test_reject_missing_field(self) -> None:
        event = {"RequestType": "Create", "ResponseURL": "https://example.com"}
        assert is_cloudformation_custom_resource(event) is False


class TestIsIotCustomAuthorizer:
    def test_match(self) -> None:
        assert is_iot_custom_authorizer({"protocolData": {}, "connectionMetadata": {}}) is True

    def test_reject(self) -> None:
        assert is_iot_custom_authorizer({"protocolData": {}}) is False


class TestIsFirehoseTransformation:
    def test_match(self) -> None:
        event = {"invocationId": "inv-1", "records": [{"recordId": "r1", "data": "Zm9v"}]}
        assert is_firehose_transformation(event) is True

    def test_reject_no_records(self) -> None:
        assert is_firehose_transformation({"invocationId": "inv-1"}) is False


class TestIsSecretsManagerRotation:
    def test_match(self) -> None:
        event = {"Step": "createSecret", "SecretId": "arn:aws:secretsmanager:...", "ClientRequestToken": "token-1"}
        assert is_secrets_manager_rotation(event) is True

    def test_reject_missing_token(self) -> None:
        event = {"Step": "createSecret", "SecretId": "arn:aws:secretsmanager:..."}
        assert is_secrets_manager_rotation(event) is False


class TestIsTransferFamilyAuth:
    def test_match(self) -> None:
        event = {"username": "u1", "serverId": "s-123", "sourceIp": "127.0.0.1"}
        assert is_transfer_family_auth(event) is True

    def test_reject(self) -> None:
        assert is_transfer_family_auth({"username": "u1"}) is False


# ---------- Edge cases ----------

class TestEdgeCases:
    """Verify all detectors gracefully reject non-object inputs."""

    @pytest.mark.parametrize("bad_input", [None, 42, "hello", [], True])
    def test_non_dict_rejected(self, bad_input: object) -> None:
        assert is_http_v1(bad_input) is False
        assert is_http_v2(bad_input) is False
        assert is_alb(bad_input) is False
        assert is_sqs(bad_input) is False
        assert is_sns(bad_input) is False
        assert is_s3(bad_input) is False
        assert is_ses(bad_input) is False
        assert is_dynamodb_stream(bad_input) is False
        assert is_kinesis_stream(bad_input) is False
        assert is_msk(bad_input) is False
        assert is_self_managed_kafka(bad_input) is False
        assert is_cloudwatch_logs(bad_input) is False
        assert is_scheduled_event(bad_input) is False
        assert is_codebuild_state(bad_input) is False
        assert is_cloudwatch_alarm(bad_input) is False
        assert is_eventbridge(bad_input) is False
        assert is_codecommit(bad_input) is False
        assert is_codepipeline(bad_input) is False
        assert is_cloudfront_request(bad_input) is False
        assert is_cloudfront_response(bad_input) is False
        assert is_guardduty(bad_input) is False
        assert is_connect_contact_flow(bad_input) is False
        assert is_s3_batch(bad_input) is False
        assert is_s3_notification(bad_input) is False
        assert is_lex(bad_input) is False
        assert is_lex_v2(bad_input) is False
        assert is_appsync_resolver(bad_input) is False
        assert is_amplify_resolver(bad_input) is False
        assert is_api_gateway_authorizer(bad_input) is False
        assert is_autoscaling_scale_in(bad_input) is False
        assert is_cloudformation_custom_resource(bad_input) is False
        assert is_iot_custom_authorizer(bad_input) is False
        assert is_firehose_transformation(bad_input) is False
        assert is_secrets_manager_rotation(bad_input) is False
        assert is_transfer_family_auth(bad_input) is False
