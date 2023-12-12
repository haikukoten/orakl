import { Logger } from 'pino'
import type { RedisClientType } from 'redis'
import {
  L2_REPORTER_REQUEST_RESPONSE_REQUEST_QUEUE_NAME,
  L2_REQUEST_RESPONSE_REQUEST_REPORTER_STATE_NAME,
  L2_REQUEST_RESPONSE_REQUEST_SERVICE_NAME
} from '../settings'
import { factory } from './factory'

export async function buildReporter(redisClient: RedisClientType, logger: Logger) {
  await factory({
    redisClient,
    stateName: L2_REQUEST_RESPONSE_REQUEST_REPORTER_STATE_NAME,
    service: L2_REQUEST_RESPONSE_REQUEST_SERVICE_NAME,
    reporterQueueName: L2_REPORTER_REQUEST_RESPONSE_REQUEST_QUEUE_NAME,
    concurrency: 1,
    delegatedFee: false,
    _logger: logger
  })
}