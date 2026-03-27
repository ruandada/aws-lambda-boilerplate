import assert from 'node:assert/strict'
import test from 'node:test'
import {
  isAlbEvent,
  isCloudWatchLogsEvent,
  isCodePipelineEvent,
  isDynamoDbStreamEvent,
  isFirehoseTransformationEvent,
  isIoTEvent,
  isKinesisStreamEvent,
  isLambdaFunctionUrlEvent,
  isMskEvent,
  isS3Event,
  isSecretsManagerRotationEvent,
  isSelfManagedKafkaEvent,
  isSesEvent,
} from './events'

test('matches lambda function URL event (api gateway v2 shape)', () => {
  const event = {
    version: '2.0',
    requestContext: {
      http: { method: 'GET', path: '/' },
      domainName: 'abc.lambda-url.us-east-1.on.aws',
    },
  }
  assert.equal(isLambdaFunctionUrlEvent(event), true)
})

test('matches alb event when requestContext.elb exists', () => {
  const event = {
    requestContext: {
      elb: {},
    },
  }
  assert.equal(isAlbEvent(event), true)
})

test('matches classic records events for s3/ses/dynamodb/kinesis', () => {
  const s3Event = { Records: [{ eventSource: 'aws:s3', eventName: 'ObjectCreated:Put' }] }
  const sesEvent = { Records: [{ eventSource: 'aws:ses', ses: {} }] }
  const dynamodbEvent = { Records: [{ eventSource: 'aws:dynamodb', eventName: 'INSERT' }] }
  const kinesisEvent = { Records: [{ eventSource: 'aws:kinesis', eventName: 'aws:kinesis:record' }] }

  assert.equal(isS3Event(s3Event), true)
  assert.equal(isSesEvent(sesEvent), true)
  assert.equal(isDynamoDbStreamEvent(dynamodbEvent), true)
  assert.equal(isKinesisStreamEvent(kinesisEvent), true)
})

test('matches cloudwatch logs event', () => {
  const event = { awslogs: { data: 'H4sIAAAAAAAA...' } }
  assert.equal(isCloudWatchLogsEvent(event), true)
})

test('matches codepipeline/firehose/secrets-manager event shapes', () => {
  const codePipelineEvent = { 'CodePipeline.job': { id: 'job-123' } }
  const firehoseEvent = { invocationId: 'inv-1', records: [{ recordId: 'r1', data: 'Zm9v' }] }
  const secretsRotationEvent = {
    Step: 'createSecret',
    SecretId: 'arn:aws:secretsmanager:...',
    ClientRequestToken: 'token-1',
  }

  assert.equal(isCodePipelineEvent(codePipelineEvent), true)
  assert.equal(isFirehoseTransformationEvent(firehoseEvent), true)
  assert.equal(isSecretsManagerRotationEvent(secretsRotationEvent), true)
})

test('matches msk and self-managed kafka events', () => {
  const mskEvent = {
    eventSource: 'aws:kafka',
    records: {
      'topic-0': [{ topic: 'topic', partition: 0, offset: 1 }],
    },
  }
  const selfManagedKafkaEvent = {
    eventSource: 'SelfManagedKafka',
    records: {
      'topic-0': [{ topic: 'topic', partition: 0, offset: 1 }],
    },
  }

  assert.equal(isMskEvent(mskEvent), true)
  assert.equal(isSelfManagedKafkaEvent(selfManagedKafkaEvent), true)
})

test('matches iot scalar payloads and rejects non-scalars', () => {
  assert.equal(isIoTEvent('temperature:42'), true)
  assert.equal(isIoTEvent(42), true)
  assert.equal(isIoTEvent([]), false)
})

test('does not misclassify unrelated payloads', () => {
  const unknown = { foo: 'bar' }
  assert.equal(isCodePipelineEvent(unknown), false)
  assert.equal(isFirehoseTransformationEvent(unknown), false)
  assert.equal(isSecretsManagerRotationEvent(unknown), false)
  assert.equal(isMskEvent(unknown), false)
  assert.equal(isSelfManagedKafkaEvent(unknown), false)
  assert.equal(isCloudWatchLogsEvent(unknown), false)
})

test('matches http v1 event', () => {
  const event = { httpMethod: 'GET', path: '/' }
  const { isApiGatewayProxyEvent } = require('./events')
  assert.equal(isApiGatewayProxyEvent(event), true)
})

test('matches http v2 event', () => {
  const event = {
    version: '2.0',
    requestContext: {
      http: { method: 'GET', path: '/' },
    },
  }
  const { isApiGatewayProxyEventV2 } = require('./events')
  assert.equal(isApiGatewayProxyEventV2(event), true)
})

test('isHttpEvent matches any http-family event', () => {
  const { isHttpEvent } = require('./events')
  assert.equal(isHttpEvent({ httpMethod: 'GET' }), true)
  assert.equal(isHttpEvent({ version: '2.0', requestContext: { http: { method: 'GET' } } }), true)
  assert.equal(isHttpEvent({ requestContext: { elb: {} } }), true)
  assert.equal(isHttpEvent({ foo: 'bar' }), false)
})

test('isSqsEvent matches sqs records', () => {
  const { isSqsEvent } = require('./events')
  assert.equal(isSqsEvent({ Records: [{ eventSource: 'aws:sqs' }] }), true)
  assert.equal(isSqsEvent({ Records: [{ eventSource: 'aws:sns' }] }), false)
})

test('isSnsEvent matches sns records', () => {
  const { isSnsEvent } = require('./events')
  assert.equal(isSnsEvent({ Records: [{ EventSource: 'aws:sns' }] }), true)
  assert.equal(isSnsEvent({ Records: [{ eventSource: 'aws:sqs' }] }), false)
})

test('isScheduledEvent matches aws.events scheduled', () => {
  const { isScheduledEvent } = require('./events')
  assert.equal(isScheduledEvent({ source: 'aws.events', 'detail-type': 'Scheduled Event' }), true)
  assert.equal(isScheduledEvent({ source: 'custom.app', 'detail-type': 'User Created' }), false)
})

test('isEventBridgeEvent with options filters correctly', () => {
  const { isEventBridgeEvent } = require('./events')
  const event = { source: 'custom.app', 'detail-type': 'OrderCreated', detail: {} }
  assert.equal(isEventBridgeEvent(event), true)
  assert.equal(isEventBridgeEvent(event, { source: 'custom.app' }), true)
  assert.equal(isEventBridgeEvent(event, { source: 'other.app' }), false)
  assert.equal(isEventBridgeEvent(event, { detailType: 'OrderCreated' }), true)
  assert.equal(isEventBridgeEvent(event, { detailType: 'Other' }), false)
})

test('createEventBridgeEventMatcher returns a reusable predicate', () => {
  const { createEventBridgeEventMatcher } = require('./events')
  const matcher = createEventBridgeEventMatcher({ source: 'custom.app' })
  assert.equal(matcher({ source: 'custom.app', 'detail-type': 'Test' }), true)
  assert.equal(matcher({ source: 'other.app', 'detail-type': 'Test' }), false)
})

test('isCodeBuildStateEvent matches codebuild state changes', () => {
  const { isCodeBuildStateEvent } = require('./events')
  assert.equal(
    isCodeBuildStateEvent({ source: 'aws.codebuild', 'detail-type': 'CodeBuild Build State Change' }),
    true,
  )
  assert.equal(isCodeBuildStateEvent({ source: 'aws.codebuild', 'detail-type': 'Other' }), false)
})

test('isCodeCommitEvent matches codecommit records', () => {
  const { isCodeCommitEvent } = require('./events')
  assert.equal(isCodeCommitEvent({ Records: [{ codecommit: { references: [] } }] }), true)
  assert.equal(isCodeCommitEvent({ Records: [{ eventSource: 'aws:sqs' }] }), false)
})

test('isCloudWatchAlarmEvent matches alarm shape', () => {
  const { isCloudWatchAlarmEvent } = require('./events')
  assert.equal(isCloudWatchAlarmEvent({ alarmArn: 'arn:aws:cloudwatch:...', alarmData: {} }), true)
  assert.equal(isCloudWatchAlarmEvent({ alarmArn: 'arn:aws:cloudwatch:...' }), false)
})

test('isConnectContactFlowEvent matches contact flow', () => {
  const { isConnectContactFlowEvent } = require('./events')
  assert.equal(isConnectContactFlowEvent({ Name: 'ContactFlowEvent', Details: { ContactData: {} } }), true)
  assert.equal(isConnectContactFlowEvent({ Name: 'Other' }), false)
})

test('isCloudFrontRequestEvent matches request without response', () => {
  const { isCloudFrontRequestEvent } = require('./events')
  assert.equal(isCloudFrontRequestEvent({ Records: [{ cf: { request: {} } }] }), true)
  assert.equal(isCloudFrontRequestEvent({ Records: [{ cf: { request: {}, response: {} } }] }), false)
})

test('isCloudFrontResponseEvent matches response', () => {
  const { isCloudFrontResponseEvent } = require('./events')
  assert.equal(isCloudFrontResponseEvent({ Records: [{ cf: { response: {} } }] }), true)
  assert.equal(isCloudFrontResponseEvent({ Records: [{ eventSource: 'aws:sqs' }] }), false)
})

test('isGuardDutyNotificationEvent matches guardduty source', () => {
  const { isGuardDutyNotificationEvent } = require('./events')
  assert.equal(isGuardDutyNotificationEvent({ source: 'aws.guardduty', 'detail-type': 'Finding' }), true)
  assert.equal(isGuardDutyNotificationEvent({ source: 'aws.events', 'detail-type': 'X' }), false)
})

test('isS3BatchEvent matches batch operations shape', () => {
  const { isS3BatchEvent } = require('./events')
  assert.equal(isS3BatchEvent({ invocationSchemaVersion: '1.0', tasks: [] }), true)
  assert.equal(isS3BatchEvent({ invocationSchemaVersion: '1.0' }), false)
})

test('isS3NotificationEvent matches s3 eventbridge notification', () => {
  const { isS3NotificationEvent } = require('./events')
  assert.equal(isS3NotificationEvent({ source: 'aws.s3', 'detail-type': 'Object Created' }), true)
  assert.equal(isS3NotificationEvent({ source: 'aws.ec2', 'detail-type': 'X' }), false)
})

test('isLexEvent and isLexV2Event match respective shapes', () => {
  const { isLexEvent, isLexV2Event } = require('./events')
  assert.equal(isLexEvent({ messageVersion: '1.0', currentIntent: {}, bot: {} }), true)
  assert.equal(isLexEvent({ messageVersion: '1.0', bot: {} }), false)
  assert.equal(isLexV2Event({ messageVersion: '1.0', bot: {}, sessionId: 'sid' }), true)
  assert.equal(isLexV2Event({ messageVersion: '1.0', bot: {} }), false)
})

test('isAppSyncResolverEvent matches appsync shape', () => {
  const { isAppSyncResolverEvent } = require('./events')
  assert.equal(isAppSyncResolverEvent({ info: { fieldName: 'listPosts' }, request: {} }), true)
  assert.equal(isAppSyncResolverEvent({ request: {} }), false)
})

test('isAmplifyResolverEvent matches amplify shape', () => {
  const { isAmplifyResolverEvent } = require('./events')
  assert.equal(isAmplifyResolverEvent({ typeName: 'Query', fieldName: 'list', request: {} }), true)
  assert.equal(isAmplifyResolverEvent({ typeName: 'Query', request: {} }), false)
})

test('isApiGatewayAuthorizerEvent matches TOKEN/REQUEST types', () => {
  const { isApiGatewayAuthorizerEvent } = require('./events')
  assert.equal(isApiGatewayAuthorizerEvent({ type: 'TOKEN', methodArn: 'arn:...' }), true)
  assert.equal(isApiGatewayAuthorizerEvent({ type: 'REQUEST', routeArn: 'arn:...' }), true)
  assert.equal(isApiGatewayAuthorizerEvent({ type: 'INVALID', methodArn: 'arn:...' }), false)
})

test('isAutoScalingScaleInEvent matches scale-in shape', () => {
  const { isAutoScalingScaleInEvent } = require('./events')
  assert.equal(
    isAutoScalingScaleInEvent({ AutoScalingGroupARN: 'arn:...', CapacityToTerminate: [] }),
    true,
  )
  assert.equal(isAutoScalingScaleInEvent({ AutoScalingGroupARN: 'arn:...' }), false)
})

test('isCloudFormationCustomResourceEvent matches all required fields', () => {
  const { isCloudFormationCustomResourceEvent } = require('./events')
  assert.equal(
    isCloudFormationCustomResourceEvent({
      RequestType: 'Create',
      ResponseURL: 'https://example.com',
      StackId: 'stack',
      RequestId: 'req',
      LogicalResourceId: 'Res',
      ResourceType: 'Custom::T',
    }),
    true,
  )
  assert.equal(
    isCloudFormationCustomResourceEvent({ RequestType: 'Create', ResponseURL: 'https://example.com' }),
    false,
  )
})

test('isIoTCustomAuthorizerEvent matches iot authorizer shape', () => {
  const { isIoTCustomAuthorizerEvent } = require('./events')
  assert.equal(isIoTCustomAuthorizerEvent({ protocolData: {}, connectionMetadata: {} }), true)
  assert.equal(isIoTCustomAuthorizerEvent({ protocolData: {} }), false)
})

test('isTransferFamilyAuthorizerEvent matches transfer family shape', () => {
  const { isTransferFamilyAuthorizerEvent } = require('./events')
  assert.equal(
    isTransferFamilyAuthorizerEvent({ username: 'u1', serverId: 's-123', sourceIp: '127.0.0.1' }),
    true,
  )
  assert.equal(isTransferFamilyAuthorizerEvent({ username: 'u1' }), false)
})
