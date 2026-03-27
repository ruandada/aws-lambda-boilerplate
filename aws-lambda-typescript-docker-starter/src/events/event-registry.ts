import type {
  ALBEvent,
  ALBResult,
  AmplifyGraphQlResolverEvent,
  APIGatewayAuthorizerEvent,
  APIGatewayAuthorizerResult,
  APIGatewayProxyEvent,
  APIGatewayProxyEventV2,
  APIGatewayProxyResult,
  APIGatewayProxyResultV2,
  AppSyncResolverEvent,
  AutoScalingScaleInEvent,
  AutoScalingScaleInResult,
  CdkCustomResourceEvent,
  CdkCustomResourceResponse,
  CloudFormationCustomResourceEvent,
  CloudFrontRequestEvent,
  CloudFrontRequestResult,
  CloudFrontResponseEvent,
  CloudFrontResponseResult,
  CloudWatchAlarmEvent,
  CloudWatchLogsEvent,
  CodeBuildCloudWatchStateEvent,
  CodeCommitTriggerEvent,
  CodePipelineEvent,
  ConnectContactFlowEvent,
  ConnectContactFlowResult,
  Context,
  DynamoDBBatchResponse,
  DynamoDBStreamEvent,
  EventBridgeEvent,
  FirehoseTransformationEvent,
  FirehoseTransformationResult,
  GuardDutyNotificationEvent,
  IoTCustomAuthorizerEvent,
  IoTCustomAuthorizerResult,
  IoTEvent,
  KinesisStreamBatchResponse,
  KinesisStreamEvent,
  LambdaFunctionURLEvent,
  LambdaFunctionURLResult,
  LexEvent,
  LexResult,
  LexV2Event,
  LexV2Result,
  MSKEvent,
  S3BatchEvent,
  S3BatchResult,
  S3Event,
  S3NotificationEvent,
  ScheduledEvent,
  SecretsManagerRotationEvent,
  SelfManagedKafkaEvent,
  SESEvent,
  SNSEvent,
  SQSBatchResponse,
  SQSEvent,
  TransferFamilyAuthorizerEvent,
  TransferFamilyAuthorizerResult,
} from 'aws-lambda'
import {
  createEventBridgeEventMatcher,
  isAlbEvent,
  isAmplifyResolverEvent,
  isApiGatewayAuthorizerEvent,
  isApiGatewayProxyEvent,
  isApiGatewayProxyEventV2,
  isAppSyncResolverEvent,
  isAutoScalingScaleInEvent,
  isCdkCustomResourceEvent,
  isCloudFormationCustomResourceEvent,
  isCloudFrontRequestEvent,
  isCloudFrontResponseEvent,
  isCloudWatchAlarmEvent,
  isCloudWatchLogsEvent,
  isCodeBuildStateEvent,
  isCodeCommitEvent,
  isCodePipelineEvent,
  isConnectContactFlowEvent,
  isDynamoDbStreamEvent,
  isFirehoseTransformationEvent,
  isGuardDutyNotificationEvent,
  isHttpEvent,
  isIoTCustomAuthorizerEvent,
  isIoTEvent,
  isKinesisStreamEvent,
  isLambdaFunctionUrlEvent,
  isLexEvent,
  isLexV2Event,
  isMskEvent,
  isS3BatchEvent,
  isS3Event,
  isS3NotificationEvent,
  isScheduledEvent,
  isSecretsManagerRotationEvent,
  isSelfManagedKafkaEvent,
  isSesEvent,
  isSnsEvent,
  isSqsEvent,
  isTransferFamilyAuthorizerEvent,
} from './events'

type EventPredicate<TEvent> = (event: unknown) => event is TEvent
type EventHandler<TEvent, TResult = void> = (event: TEvent, context: Context) => Promise<TResult>

interface EventRegistration<TEvent = unknown, TResult = void> {
  match: EventPredicate<TEvent>
  handle: EventHandler<TEvent, TResult>
}

export type DispatchResult<TResult = unknown> = { matched: true; result: TResult } | { matched: false }

export class EventRegistry {
  private readonly registrations: EventRegistration<any, any>[] = []

  /** Register a custom event matcher and handler. Supports method chaining. */
  register<TEvent, TResult>(match: EventPredicate<TEvent>, handle: EventHandler<TEvent, TResult>): this {
    this.registrations.push({ match, handle })
    return this
  }

  /** Dispatch the event to the first matching handler; returns matched=false when none match. */
  async dispatch(event: unknown, context: Context): Promise<DispatchResult> {
    for (const registration of this.registrations) {
      if (registration.match(event)) {
        const result = await registration.handle(event, context)
        return { matched: true, result }
      }
    }
    return { matched: false }
  }

  /** Register an HTTP handler for Amazon API Gateway (v1/v2) and Elastic Load Balancing (ALB). */
  registerHttpEvent<TResult = APIGatewayProxyResult | APIGatewayProxyResultV2 | ALBResult>(
    handle: EventHandler<APIGatewayProxyEvent | APIGatewayProxyEventV2 | ALBEvent, TResult>,
  ): this {
    return this.register(isHttpEvent, handle)
  }

  /** Register a proxy handler for Amazon API Gateway REST API / HTTP API payload v1 events. */
  registerApiGatewayProxyEvent<TResult = APIGatewayProxyResult>(
    handle: EventHandler<APIGatewayProxyEvent, TResult>,
  ): this {
    return this.register(isApiGatewayProxyEvent, handle)
  }

  /** Register a proxy handler for Amazon API Gateway HTTP API payload v2 events. */
  registerApiGatewayProxyEventV2<TResult = APIGatewayProxyResultV2>(
    handle: EventHandler<APIGatewayProxyEventV2, TResult>,
  ): this {
    return this.register(isApiGatewayProxyEventV2, handle)
  }

  /** Register a handler for Elastic Load Balancing (ALB) Lambda target events. */
  registerAlbEvent<TResult = ALBResult>(handle: EventHandler<ALBEvent, TResult>): this {
    return this.register(isAlbEvent, handle)
  }

  /** Register a handler for AWS Lambda Function URL invocation events. */
  registerLambdaFunctionUrlEvent<TResult = LambdaFunctionURLResult>(
    handle: EventHandler<LambdaFunctionURLEvent, TResult>,
  ): this {
    return this.register(isLambdaFunctionUrlEvent, handle)
  }

  /** Register a handler for Amazon SQS queue events. */
  registerSqsEvent<TResult = SQSBatchResponse | void>(handle: EventHandler<SQSEvent, TResult>): this {
    return this.register(isSqsEvent, handle)
  }

  /** Register a handler for Amazon SNS notification events. */
  registerSnsEvent<TResult = void>(handle: EventHandler<SNSEvent, TResult>): this {
    return this.register(isSnsEvent, handle)
  }

  /** Register a handler for Amazon S3 notification events (classic Records format). */
  registerS3Event<TResult = void>(handle: EventHandler<S3Event, TResult>): this {
    return this.register(isS3Event, handle)
  }

  /** Register a handler for Amazon SES inbound mail events. */
  registerSesEvent<TResult = void>(handle: EventHandler<SESEvent, TResult>): this {
    return this.register(isSesEvent, handle)
  }

  /** Register a handler for Amazon DynamoDB Streams events. */
  registerDynamoDbStreamEvent<TResult = DynamoDBBatchResponse | void>(
    handle: EventHandler<DynamoDBStreamEvent, TResult>,
  ): this {
    return this.register(isDynamoDbStreamEvent, handle)
  }

  /** Register a handler for Amazon Kinesis Data Streams events. */
  registerKinesisStreamEvent<TResult = KinesisStreamBatchResponse | void>(
    handle: EventHandler<KinesisStreamEvent, TResult>,
  ): this {
    return this.register(isKinesisStreamEvent, handle)
  }

  /** Register a handler for Amazon MSK (Managed Kafka) events. */
  registerMskEvent<TResult = void>(handle: EventHandler<MSKEvent, TResult>): this {
    return this.register(isMskEvent, handle)
  }

  /** Register a handler for self-managed Kafka events integrated with AWS Lambda. */
  registerSelfManagedKafkaEvent<TResult = void>(handle: EventHandler<SelfManagedKafkaEvent, TResult>): this {
    return this.register(isSelfManagedKafkaEvent, handle)
  }

  /** Register a handler for Amazon CloudWatch Logs subscription events. */
  registerCloudWatchLogsEvent<TResult = void>(handle: EventHandler<CloudWatchLogsEvent, TResult>): this {
    return this.register(isCloudWatchLogsEvent, handle)
  }

  /** Register a generic Amazon EventBridge handler with optional source/detail-type filters. */
  registerEventBridgeEvent<TDetailType extends string, TDetail, TResult = void>(
    handle: EventHandler<EventBridgeEvent<TDetailType, TDetail>, TResult>,
    options?: { source?: string; detailType?: TDetailType },
  ): this {
    return this.register(createEventBridgeEventMatcher<TDetailType, TDetail>(options), handle)
  }

  /** Register a handler for Amazon EventBridge scheduled events. */
  registerScheduledEvent<TDetail, TResult = void>(handle: EventHandler<ScheduledEvent<TDetail>, TResult>): this {
    return this.register(isScheduledEvent<TDetail>, handle)
  }

  /** Register a handler for AWS CodeBuild build state change events (via EventBridge). */
  registerCodeBuildStateEvent<TResult = void>(handle: EventHandler<CodeBuildCloudWatchStateEvent, TResult>): this {
    return this.register(isCodeBuildStateEvent, handle)
  }

  /** Register a handler for AWS CodeCommit repository trigger events. */
  registerCodeCommitEvent<TResult = void>(handle: EventHandler<CodeCommitTriggerEvent, TResult>): this {
    return this.register(isCodeCommitEvent, handle)
  }

  /** Register a handler for AWS CodePipeline job invocation events. */
  registerCodePipelineEvent<TResult = void>(handle: EventHandler<CodePipelineEvent, TResult>): this {
    return this.register(isCodePipelineEvent, handle)
  }

  /** Register a handler for Amazon CloudWatch Alarm state change events. */
  registerCloudWatchAlarmEvent<TResult = void>(handle: EventHandler<CloudWatchAlarmEvent, TResult>): this {
    return this.register(isCloudWatchAlarmEvent, handle)
  }

  /** Register a handler for Amazon Connect contact flow events. */
  registerConnectContactFlowEvent<TResult = ConnectContactFlowResult>(
    handle: EventHandler<ConnectContactFlowEvent, TResult>,
  ): this {
    return this.register(isConnectContactFlowEvent, handle)
  }

  /** Register a handler for Amazon CloudFront request events. */
  registerCloudFrontRequestEvent<TResult = CloudFrontRequestResult>(
    handle: EventHandler<CloudFrontRequestEvent, TResult>,
  ): this {
    return this.register(isCloudFrontRequestEvent, handle)
  }

  /** Register a handler for Amazon CloudFront response events. */
  registerCloudFrontResponseEvent<TResult = CloudFrontResponseResult>(
    handle: EventHandler<CloudFrontResponseEvent, TResult>,
  ): this {
    return this.register(isCloudFrontResponseEvent, handle)
  }

  /** Register a handler for Amazon GuardDuty notification events (via EventBridge). */
  registerGuardDutyNotificationEvent<TResult = void>(handle: EventHandler<GuardDutyNotificationEvent, TResult>): this {
    return this.register(isGuardDutyNotificationEvent, handle)
  }

  /** Register a handler for Amazon S3 Batch Operations events. */
  registerS3BatchEvent<TResult = S3BatchResult>(handle: EventHandler<S3BatchEvent, TResult>): this {
    return this.register(isS3BatchEvent, handle)
  }

  /** Register a handler for Amazon S3 notifications delivered through EventBridge. */
  registerS3NotificationEvent<TResult = void>(handle: EventHandler<S3NotificationEvent, TResult>): this {
    return this.register(isS3NotificationEvent, handle)
  }

  /** Register a handler for Amazon Lex V1 bot events. */
  registerLexEvent<TResult = LexResult>(handle: EventHandler<LexEvent, TResult>): this {
    return this.register(isLexEvent, handle)
  }

  /** Register a handler for Amazon Lex V2 bot events. */
  registerLexV2Event<TResult = LexV2Result>(handle: EventHandler<LexV2Event, TResult>): this {
    return this.register(isLexV2Event, handle)
  }

  /** Register a handler for AWS AppSync resolver invocation events. */
  registerAppSyncResolverEvent<
    TArgs = Record<string, any>,
    TSource = Record<string, any> | null,
    TResult = Record<string, any>,
  >(handle: EventHandler<AppSyncResolverEvent<TArgs, TSource>, TResult>): this {
    return this.register(isAppSyncResolverEvent<TArgs, TSource>, handle)
  }

  /** Register a handler for AWS Amplify GraphQL resolver events. */
  registerAmplifyResolverEvent<
    TArgs = Record<string, any>,
    TSource = Record<string, any>,
    TResult = Record<string, any>,
  >(handle: EventHandler<AmplifyGraphQlResolverEvent<TArgs, TSource>, TResult>): this {
    return this.register(isAmplifyResolverEvent<TArgs, TSource>, handle)
  }

  /** Register a handler for Amazon API Gateway authorizer events (TOKEN/REQUEST). */
  registerApiGatewayAuthorizerEvent<TResult = APIGatewayAuthorizerResult>(
    handle: EventHandler<APIGatewayAuthorizerEvent, TResult>,
  ): this {
    return this.register(isApiGatewayAuthorizerEvent, handle)
  }

  /** Register a handler for Amazon EC2 Auto Scaling scale-in events. */
  registerAutoScalingScaleInEvent<TResult = AutoScalingScaleInResult>(
    handle: EventHandler<AutoScalingScaleInEvent, TResult>,
  ): this {
    return this.register(isAutoScalingScaleInEvent, handle)
  }

  /** Register a handler for AWS CloudFormation custom resource lifecycle events. */
  registerCloudFormationCustomResourceEvent<TProps = Record<string, any>, TOldProps = TProps, TResult = void>(
    handle: EventHandler<CloudFormationCustomResourceEvent<TProps, TOldProps>, TResult>,
  ): this {
    return this.register(isCloudFormationCustomResourceEvent<TProps, TOldProps>, handle)
  }

  /** Register a handler for AWS CDK custom resource events. */
  registerCdkCustomResourceEvent<
    TProps = Record<string, any>,
    TOldProps = TProps,
    TResult = CdkCustomResourceResponse<Record<string, any>>,
  >(handle: EventHandler<CdkCustomResourceEvent<TProps, TOldProps>, TResult>): this {
    return this.register(isCdkCustomResourceEvent<TProps, TOldProps>, handle)
  }

  /** Register a handler for AWS IoT payload events (string or number payload). */
  registerIoTEvent<TPayload = never, TResult = void>(handle: EventHandler<IoTEvent<TPayload>, TResult>): this {
    return this.register(isIoTEvent<TPayload>, handle)
  }

  /** Register a handler for AWS IoT Core custom authorizer events. */
  registerIoTCustomAuthorizerEvent<TResult = IoTCustomAuthorizerResult>(
    handle: EventHandler<IoTCustomAuthorizerEvent, TResult>,
  ): this {
    return this.register(isIoTCustomAuthorizerEvent, handle)
  }

  /** Register a handler for Amazon Kinesis Data Firehose transformation events. */
  registerFirehoseTransformationEvent<TResult = FirehoseTransformationResult>(
    handle: EventHandler<FirehoseTransformationEvent, TResult>,
  ): this {
    return this.register(isFirehoseTransformationEvent, handle)
  }

  /** Register a handler for AWS Secrets Manager rotation lifecycle events. */
  registerSecretsManagerRotationEvent<TResult = void>(
    handle: EventHandler<SecretsManagerRotationEvent, TResult>,
  ): this {
    return this.register(isSecretsManagerRotationEvent, handle)
  }

  /** Register a handler for AWS Transfer Family custom authorizer events. */
  registerTransferFamilyAuthorizerEvent<TResult = TransferFamilyAuthorizerResult>(
    handle: EventHandler<TransferFamilyAuthorizerEvent, TResult>,
  ): this {
    return this.register(isTransferFamilyAuthorizerEvent, handle)
  }
}
