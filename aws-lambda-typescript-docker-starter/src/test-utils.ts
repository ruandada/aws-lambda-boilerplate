import type { APIGatewayProxyEventV2, Context } from 'aws-lambda'

export function createContext(): Context {
  return {
    callbackWaitsForEmptyEventLoop: false,
    functionName: 'local-test-function',
    functionVersion: '$LATEST',
    invokedFunctionArn: 'arn:aws:lambda:us-east-1:123456789012:function:local-test-function',
    memoryLimitInMB: '128',
    awsRequestId: 'test-request-id',
    logGroupName: '/aws/lambda/local-test-function',
    logStreamName: '2026/03/24/[$LATEST]local',
    getRemainingTimeInMillis: () => 30000,
    done: () => undefined,
    fail: () => undefined,
    succeed: () => undefined,
  }
}

export function createApiGatewayV2Event(path = '/'): APIGatewayProxyEventV2 {
  return {
    version: '2.0',
    routeKey: '$default',
    rawPath: path,
    rawQueryString: '',
    headers: {
      host: 'localhost',
      'content-type': 'application/json',
    },
    requestContext: {
      accountId: '123456789012',
      apiId: 'local-api',
      domainName: 'localhost',
      domainPrefix: 'localhost',
      requestId: 'request-id',
      routeKey: '$default',
      stage: '$default',
      time: '24/Mar/2026:00:00:00 +0000',
      timeEpoch: 0,
      http: {
        method: 'GET',
        path,
        protocol: 'HTTP/1.1',
        sourceIp: '127.0.0.1',
        userAgent: 'node-test',
      },
    },
    isBase64Encoded: false,
  }
}
