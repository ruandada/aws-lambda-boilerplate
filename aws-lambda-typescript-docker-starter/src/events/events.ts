import type {
  ALBEvent,
  AmplifyGraphQlResolverEvent,
  APIGatewayAuthorizerEvent,
  APIGatewayProxyEvent,
  APIGatewayProxyEventV2,
  AppSyncResolverEvent,
  AutoScalingScaleInEvent,
  CdkCustomResourceEvent,
  CloudFormationCustomResourceEvent,
  CloudFrontRequestEvent,
  CloudFrontResponseEvent,
  CloudWatchAlarmEvent,
  CloudWatchLogsEvent,
  CodeBuildCloudWatchStateEvent,
  CodeCommitTriggerEvent,
  CodePipelineEvent,
  ConnectContactFlowEvent,
  DynamoDBStreamEvent,
  EventBridgeEvent,
  FirehoseTransformationEvent,
  GuardDutyNotificationEvent,
  IoTCustomAuthorizerEvent,
  IoTEvent,
  KinesisStreamEvent,
  LambdaFunctionURLEvent,
  LexEvent,
  LexV2Event,
  MSKEvent,
  S3BatchEvent,
  S3Event,
  S3NotificationEvent,
  ScheduledEvent,
  SESEvent,
  SelfManagedKafkaEvent,
  SNSEvent,
  SQSEvent,
  SecretsManagerRotationEvent,
  TransferFamilyAuthorizerEvent,
} from 'aws-lambda'

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function hasString(value: Record<string, unknown>, key: string): boolean {
  return typeof value[key] === 'string'
}

function hasRecordSource(
  event: unknown,
  expectedSource: string,
  sourceKeys: ReadonlyArray<'eventSource' | 'EventSource'> = ['eventSource', 'EventSource'],
): boolean {
  if (!isObject(event) || !Array.isArray(event.Records) || event.Records.length === 0) {
    return false
  }

  const first = event.Records[0]
  if (!isObject(first)) {
    return false
  }

  return sourceKeys.some((key) => first[key] === expectedSource)
}

export function isApiGatewayProxyEvent(event: unknown): event is APIGatewayProxyEvent {
  return isObject(event) && typeof event.httpMethod === 'string'
}

export function isApiGatewayProxyEventV2(event: unknown): event is APIGatewayProxyEventV2 {
  if (!isObject(event) || event.version !== '2.0' || !isObject(event.requestContext)) {
    return false
  }
  const http = event.requestContext.http
  return isObject(http) && typeof http.method === 'string'
}

export function isAlbEvent(event: unknown): event is ALBEvent {
  return isObject(event) && isObject(event.requestContext) && isObject(event.requestContext.elb)
}

export function isHttpEvent(event: unknown): event is APIGatewayProxyEvent | APIGatewayProxyEventV2 | ALBEvent {
  return isAlbEvent(event) || isApiGatewayProxyEvent(event) || isApiGatewayProxyEventV2(event)
}

export function isLambdaFunctionUrlEvent(event: unknown): event is LambdaFunctionURLEvent {
  return isApiGatewayProxyEventV2(event)
}

export function isSqsEvent(event: unknown): event is SQSEvent {
  return hasRecordSource(event, 'aws:sqs', ['eventSource'])
}

export function isSnsEvent(event: unknown): event is SNSEvent {
  return hasRecordSource(event, 'aws:sns', ['EventSource'])
}

export function isS3Event(event: unknown): event is S3Event {
  return hasRecordSource(event, 'aws:s3')
}

export function isSesEvent(event: unknown): event is SESEvent {
  return hasRecordSource(event, 'aws:ses', ['eventSource'])
}

export function isDynamoDbStreamEvent(event: unknown): event is DynamoDBStreamEvent {
  return hasRecordSource(event, 'aws:dynamodb', ['eventSource'])
}

export function isKinesisStreamEvent(event: unknown): event is KinesisStreamEvent {
  return hasRecordSource(event, 'aws:kinesis', ['eventSource'])
}

export function isMskEvent(event: unknown): event is MSKEvent {
  return isObject(event) && event.eventSource === 'aws:kafka' && isObject(event.records)
}

export function isSelfManagedKafkaEvent(event: unknown): event is SelfManagedKafkaEvent {
  return isObject(event) && event.eventSource === 'SelfManagedKafka' && isObject(event.records)
}

export function isCloudWatchLogsEvent(event: unknown): event is CloudWatchLogsEvent {
  return isObject(event) && isObject(event.awslogs) && typeof event.awslogs.data === 'string'
}

export function isEventBridgeEnvelope(event: unknown): event is EventBridgeEvent<string, unknown> {
  return isObject(event) && hasString(event, 'source') && hasString(event, 'detail-type')
}

export function isEventBridgeEvent<TDetailType extends string, TDetail>(
  event: unknown,
  options?: { source?: string; detailType?: TDetailType },
): event is EventBridgeEvent<TDetailType, TDetail> {
  if (!isEventBridgeEnvelope(event)) {
    return false
  }
  if (options?.source && event.source !== options.source) {
    return false
  }
  if (options?.detailType && event['detail-type'] !== options.detailType) {
    return false
  }
  return true
}

export function createEventBridgeEventMatcher<TDetailType extends string, TDetail>(
  options?: { source?: string; detailType?: TDetailType },
): (event: unknown) => event is EventBridgeEvent<TDetailType, TDetail> {
  return (event: unknown): event is EventBridgeEvent<TDetailType, TDetail> =>
    isEventBridgeEvent<TDetailType, TDetail>(event, options)
}

export function isScheduledEvent<TDetail>(event: unknown): event is ScheduledEvent<TDetail> {
  return isEventBridgeEnvelope(event) && event.source === 'aws.events' && event['detail-type'] === 'Scheduled Event'
}

export function isCodeBuildStateEvent(event: unknown): event is CodeBuildCloudWatchStateEvent {
  return (
    isEventBridgeEnvelope(event) &&
    event.source === 'aws.codebuild' &&
    event['detail-type'] === 'CodeBuild Build State Change'
  )
}

export function isCodeCommitEvent(event: unknown): event is CodeCommitTriggerEvent {
  return (
    isObject(event) &&
    Array.isArray(event.Records) &&
    event.Records.length > 0 &&
    isObject(event.Records[0]) &&
    isObject(event.Records[0].codecommit)
  )
}

export function isCodePipelineEvent(event: unknown): event is CodePipelineEvent {
  return isObject(event) && isObject(event['CodePipeline.job'])
}

export function isCloudWatchAlarmEvent(event: unknown): event is CloudWatchAlarmEvent {
  return isObject(event) && hasString(event, 'alarmArn') && isObject(event.alarmData)
}

export function isConnectContactFlowEvent(event: unknown): event is ConnectContactFlowEvent {
  return (
    isObject(event) &&
    event.Name === 'ContactFlowEvent' &&
    isObject(event.Details) &&
    isObject(event.Details.ContactData)
  )
}

export function isCloudFrontRequestEvent(event: unknown): event is CloudFrontRequestEvent {
  return (
    isObject(event) &&
    Array.isArray(event.Records) &&
    event.Records.length > 0 &&
    isObject(event.Records[0]) &&
    isObject(event.Records[0].cf) &&
    isObject(event.Records[0].cf.request) &&
    !isObject(event.Records[0].cf.response)
  )
}

export function isCloudFrontResponseEvent(event: unknown): event is CloudFrontResponseEvent {
  return (
    isObject(event) &&
    Array.isArray(event.Records) &&
    event.Records.length > 0 &&
    isObject(event.Records[0]) &&
    isObject(event.Records[0].cf) &&
    isObject(event.Records[0].cf.response)
  )
}

export function isGuardDutyNotificationEvent(event: unknown): event is GuardDutyNotificationEvent {
  return isEventBridgeEnvelope(event) && event.source === 'aws.guardduty'
}

export function isS3BatchEvent(event: unknown): event is S3BatchEvent {
  return isObject(event) && hasString(event, 'invocationSchemaVersion') && Array.isArray(event.tasks)
}

export function isS3NotificationEvent(event: unknown): event is S3NotificationEvent {
  return isEventBridgeEnvelope(event) && event.source === 'aws.s3'
}

export function isLexEvent(event: unknown): event is LexEvent {
  return isObject(event) && hasString(event, 'messageVersion') && isObject(event.currentIntent) && isObject(event.bot)
}

export function isLexV2Event(event: unknown): event is LexV2Event {
  return isObject(event) && hasString(event, 'messageVersion') && isObject(event.bot) && hasString(event, 'sessionId')
}

export function isAppSyncResolverEvent<TArgs = Record<string, unknown>, TSource = Record<string, unknown> | null>(
  event: unknown,
): event is AppSyncResolverEvent<TArgs, TSource> {
  return isObject(event) && isObject(event.info) && hasString(event.info, 'fieldName') && isObject(event.request)
}

export function isAmplifyResolverEvent<TArgs = Record<string, unknown>, TSource = Record<string, unknown>>(
  event: unknown,
): event is AmplifyGraphQlResolverEvent<TArgs, TSource> {
  return isObject(event) && hasString(event, 'typeName') && hasString(event, 'fieldName') && isObject(event.request)
}

export function isApiGatewayAuthorizerEvent(event: unknown): event is APIGatewayAuthorizerEvent {
  return (
    isObject(event) &&
    (event.type === 'TOKEN' || event.type === 'REQUEST') &&
    (hasString(event, 'methodArn') || hasString(event, 'routeArn'))
  )
}

export function isAutoScalingScaleInEvent(event: unknown): event is AutoScalingScaleInEvent {
  return isObject(event) && hasString(event, 'AutoScalingGroupARN') && Array.isArray(event.CapacityToTerminate)
}

export function isCloudFormationCustomResourceEvent<TProps = Record<string, unknown>, TOldProps = TProps>(
  event: unknown,
): event is CloudFormationCustomResourceEvent<TProps, TOldProps> {
  return (
    isObject(event) &&
    hasString(event, 'RequestType') &&
    hasString(event, 'ResponseURL') &&
    hasString(event, 'StackId') &&
    hasString(event, 'RequestId') &&
    hasString(event, 'LogicalResourceId') &&
    hasString(event, 'ResourceType')
  )
}

export function isCdkCustomResourceEvent<TProps = Record<string, unknown>, TOldProps = TProps>(
  event: unknown,
): event is CdkCustomResourceEvent<TProps, TOldProps> {
  return isCloudFormationCustomResourceEvent<TProps, TOldProps>(event)
}

export function isIoTEvent<TPayload = never>(event: unknown): event is IoTEvent<TPayload> {
  return typeof event === 'string' || typeof event === 'number'
}

export function isIoTCustomAuthorizerEvent(event: unknown): event is IoTCustomAuthorizerEvent {
  return isObject(event) && isObject(event.protocolData) && isObject(event.connectionMetadata)
}

export function isFirehoseTransformationEvent(event: unknown): event is FirehoseTransformationEvent {
  return isObject(event) && hasString(event, 'invocationId') && Array.isArray(event.records)
}

export function isSecretsManagerRotationEvent(event: unknown): event is SecretsManagerRotationEvent {
  return (
    isObject(event) &&
    hasString(event, 'Step') &&
    hasString(event, 'SecretId') &&
    hasString(event, 'ClientRequestToken')
  )
}

export function isTransferFamilyAuthorizerEvent(event: unknown): event is TransferFamilyAuthorizerEvent {
  return isObject(event) && hasString(event, 'username') && hasString(event, 'serverId') && hasString(event, 'sourceIp')
}
