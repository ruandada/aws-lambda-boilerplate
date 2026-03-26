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
