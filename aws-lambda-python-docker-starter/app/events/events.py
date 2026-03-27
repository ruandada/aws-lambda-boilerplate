from __future__ import annotations

from collections.abc import Callable, Mapping
from typing import Any

from aws_lambda_typing import events as aws_events

try:
    from typing import TypeGuard
except ImportError:
    from typing_extensions import TypeGuard


def _is_object(value: Any) -> TypeGuard[Mapping[str, Any]]:
    return isinstance(value, dict)


def _has_string(obj: Mapping[str, Any], key: str) -> bool:
    return isinstance(obj.get(key), str)


def _nested_object(root: Mapping[str, Any], *keys: str) -> Mapping[str, Any] | None:
    current: Any = root
    for key in keys:
        if not _is_object(current):
            return None
        current = current.get(key)
    return current if _is_object(current) else None


def _first_record(envelope: Mapping[str, Any]) -> Mapping[str, Any] | None:
    records = envelope.get("Records")
    if not isinstance(records, list) or not records:
        return None
    first = records[0]
    return first if _is_object(first) else None


def _first_record_field(envelope: Mapping[str, Any], key: str) -> str:
    rec = _first_record(envelope)
    if rec is None:
        return ""
    val = rec.get(key)
    return val if isinstance(val, str) else ""


def _first_record_cf(envelope: Mapping[str, Any]) -> Mapping[str, Any] | None:
    rec = _first_record(envelope)
    if rec is None:
        return None
    cf = rec.get("cf")
    return cf if _is_object(cf) else None


# --- HTTP-family events ---

def is_http_v1(event: Any) -> TypeGuard[aws_events.APIGatewayProxyEventV1]:
    return _is_object(event) and isinstance(event.get("httpMethod"), str)


def is_http_v2(event: Any) -> TypeGuard[aws_events.APIGatewayProxyEventV2]:
    if not _is_object(event) or event.get("version") != "2.0":
        return False
    http = _nested_object(event, "requestContext", "http")
    return http is not None and isinstance(http.get("method"), str)


def is_lambda_function_url(event: Any) -> bool:
    if not is_http_v2(event):
        return False
    rc = _nested_object(event, "requestContext")
    if rc is None:
        return False
    domain = rc.get("domainName", "")
    return isinstance(domain, str) and "lambda-url" in domain


def is_alb(event: Any) -> TypeGuard[aws_events.ALBEvent]:
    return _is_object(event) and _nested_object(event, "requestContext", "elb") is not None


def is_http_event(event: Any) -> bool:
    return is_alb(event) or is_http_v1(event) or is_http_v2(event)


# --- Queue / stream / storage events ---

def is_sqs(event: Any) -> TypeGuard[aws_events.SQSEvent]:
    return _first_record_field(event, "eventSource") == "aws:sqs" if _is_object(event) else False


def is_sns(event: Any) -> TypeGuard[aws_events.SNSEvent]:
    if not _is_object(event):
        return False
    return _first_record_field(event, "EventSource") == "aws:sns"


def is_s3(event: Any) -> TypeGuard[aws_events.S3Event]:
    if not _is_object(event):
        return False
    return _first_record_field(event, "eventSource") == "aws:s3"


def is_ses(event: Any) -> TypeGuard[aws_events.SESEvent]:
    if not _is_object(event):
        return False
    return _first_record_field(event, "eventSource") == "aws:ses"


def is_dynamodb_stream(event: Any) -> TypeGuard[aws_events.DynamoDBStreamEvent]:
    if not _is_object(event):
        return False
    return _first_record_field(event, "eventSource") == "aws:dynamodb"


def is_kinesis_stream(event: Any) -> TypeGuard[aws_events.KinesisStreamEvent]:
    if not _is_object(event):
        return False
    return _first_record_field(event, "eventSource") == "aws:kinesis"


def is_msk(event: Any) -> TypeGuard[aws_events.MSKEvent]:
    return _is_object(event) and event.get("eventSource") == "aws:kafka" and _is_object(event.get("records"))


def is_self_managed_kafka(event: Any) -> bool:
    return _is_object(event) and event.get("eventSource") == "SelfManagedKafka" and _is_object(event.get("records"))


# --- CloudWatch / EventBridge ---

def is_cloudwatch_logs(event: Any) -> TypeGuard[aws_events.CloudWatchLogsEvent]:
    if not _is_object(event):
        return False
    logs = _nested_object(event, "awslogs")
    return logs is not None and isinstance(logs.get("data"), str) and logs["data"].strip() != ""


def is_scheduled_event(event: Any) -> bool:
    return (
        _is_object(event)
        and event.get("source") == "aws.events"
        and event.get("detail-type") == "Scheduled Event"
    )


def is_codebuild_state(event: Any) -> bool:
    return (
        _is_object(event)
        and event.get("source") == "aws.codebuild"
        and event.get("detail-type") == "CodeBuild Build State Change"
    )


def is_cloudwatch_alarm(event: Any) -> bool:
    return _is_object(event) and _has_string(event, "alarmArn") and _is_object(event.get("alarmData"))


def is_eventbridge(event: Any) -> TypeGuard[aws_events.EventBridgeEvent]:
    return (
        _is_object(event)
        and isinstance(event.get("source"), str)
        and event["source"].strip() != ""
        and isinstance(event.get("detail-type"), str)
        and event["detail-type"].strip() != ""
    )


def create_eventbridge_matcher(
    source: str | None = None, detail_type: str | None = None
) -> Callable[[Any], bool]:
    def matcher(event: Any) -> bool:
        if not is_eventbridge(event):
            return False
        if source is not None and event.get("source") != source:
            return False
        if detail_type is not None and event.get("detail-type") != detail_type:
            return False
        return True

    return matcher


# --- Code* events ---

def is_codecommit(event: Any) -> TypeGuard[aws_events.CodeCommitMessageEvent]:
    if not _is_object(event):
        return False
    first = _first_record(event)
    return first is not None and _is_object(first.get("codecommit"))


def is_codepipeline(event: Any) -> TypeGuard[aws_events.CodePipelineEvent]:
    if not _is_object(event):
        return False
    job = _nested_object(event, "CodePipeline.job")
    return job is not None and isinstance(job.get("id"), str) and job["id"].strip() != ""


# --- CloudFront ---

def is_cloudfront_request(event: Any) -> bool:
    if not _is_object(event):
        return False
    cf = _first_record_cf(event)
    if cf is None:
        return False
    return _is_object(cf.get("request")) and not _is_object(cf.get("response"))


def is_cloudfront_response(event: Any) -> bool:
    if not _is_object(event):
        return False
    cf = _first_record_cf(event)
    return cf is not None and _is_object(cf.get("response"))


# --- Security ---

def is_guardduty(event: Any) -> bool:
    return _is_object(event) and event.get("source") == "aws.guardduty"


def is_connect_contact_flow(event: Any) -> bool:
    return (
        _is_object(event)
        and event.get("Name") == "ContactFlowEvent"
        and _nested_object(event, "Details", "ContactData") is not None
    )


# --- AI / resolver / API auth ---

def is_s3_batch(event: Any) -> TypeGuard[aws_events.S3BatchEvent]:
    return _is_object(event) and _has_string(event, "invocationSchemaVersion") and isinstance(event.get("tasks"), list)


def is_s3_notification(event: Any) -> bool:
    return (
        _is_object(event)
        and event.get("source") == "aws.s3"
        and isinstance(event.get("detail-type"), str)
        and event["detail-type"].strip() != ""
    )


def is_lex(event: Any) -> bool:
    return (
        _is_object(event)
        and _has_string(event, "messageVersion")
        and _is_object(event.get("currentIntent"))
        and _is_object(event.get("bot"))
    )


def is_lex_v2(event: Any) -> bool:
    return (
        _is_object(event)
        and _has_string(event, "messageVersion")
        and _is_object(event.get("bot"))
        and _has_string(event, "sessionId")
    )


def is_appsync_resolver(event: Any) -> TypeGuard[aws_events.AppSyncResolverEvent]:
    if not _is_object(event):
        return False
    info = _nested_object(event, "info")
    return info is not None and _has_string(info, "fieldName") and _is_object(event.get("request"))


def is_amplify_resolver(event: Any) -> bool:
    return (
        _is_object(event)
        and _has_string(event, "typeName")
        and _has_string(event, "fieldName")
        and _is_object(event.get("request"))
    )


def is_api_gateway_authorizer(event: Any) -> bool:
    if not _is_object(event):
        return False
    event_type = event.get("type")
    if event_type not in ("TOKEN", "REQUEST"):
        return False
    return _has_string(event, "methodArn") or _has_string(event, "routeArn")


# --- Infrastructure / device events ---

def is_autoscaling_scale_in(event: Any) -> bool:
    return _is_object(event) and _has_string(event, "AutoScalingGroupARN") and isinstance(event.get("CapacityToTerminate"), list)


def is_cloudformation_custom_resource(event: Any) -> TypeGuard[aws_events.CloudFormationCustomResourceEvent]:
    if not _is_object(event):
        return False
    required_keys = ("RequestType", "ResponseURL", "StackId", "RequestId", "LogicalResourceId", "ResourceType")
    return all(_has_string(event, key) for key in required_keys)


def is_iot_custom_authorizer(event: Any) -> bool:
    return _is_object(event) and _is_object(event.get("protocolData")) and _is_object(event.get("connectionMetadata"))


def is_firehose_transformation(event: Any) -> bool:
    return _is_object(event) and _has_string(event, "invocationId") and isinstance(event.get("records"), list)


def is_secrets_manager_rotation(event: Any) -> TypeGuard[aws_events.SecretsManagerRotationEvent]:
    return _is_object(event) and _has_string(event, "Step") and _has_string(event, "SecretId") and _has_string(event, "ClientRequestToken")


def is_transfer_family_auth(event: Any) -> bool:
    return _is_object(event) and _has_string(event, "username") and _has_string(event, "serverId") and _has_string(event, "sourceIp")
