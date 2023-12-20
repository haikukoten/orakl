import { Job, Queue, UnrecoverableError } from 'bullmq'
import { Logger } from 'pino'
import { OraklError, OraklErrorCode } from '../errors'
import { BULLMQ_CONNECTION, REPORTER_AGGREGATOR_QUEUE_NAME } from '../settings'
import { ITransactionParameters } from '../types'
import { isRoundIdFresh } from '../utils'
import { State } from './state'
import { wrapperType } from './types'
import { sendTransaction, sendTransactionCaver, sendTransactionDelegatedFee } from './utils'

export function reporter(state: State, logger: Logger): wrapperType {
  async function wrapper(job: Job) {
    const inData: ITransactionParameters = job.data
    logger.debug(inData, 'inData')

    const { payload, gasLimit, to } = inData

    const wallet = state.wallets[to]
    if (!wallet) {
      const msg = `Wallet for oracle ${to} is not active`
      logger.error(msg)
      throw new OraklError(OraklErrorCode.WalletNotActive, msg)
    }

    let delegatorOkay = true
    const NUM_TRANSACTION_TRIALS = 3
    const txParams = { wallet, to, payload, gasLimit, logger }

    for (let i = 0; i < NUM_TRANSACTION_TRIALS; ++i) {
      if (state.delegatedFee && delegatorOkay) {
        try {
          await sendTransactionDelegatedFee(txParams)
          break
        } catch (e) {
          if (e.code == OraklErrorCode.DelegatorServerIssue) {
            delegatorOkay = false
          }
        }
      } else if (state.delegatedFee) {
        try {
          await sendTransactionCaver(txParams)
          break
        } catch (e) {
          if (![OraklErrorCode.CaverTxTransactionFailed].includes(e.code)) {
            throw e
          }
        }
      } else {
        try {
          await sendTransaction(txParams)
          break
        } catch (e) {
          if (
            ![
              OraklErrorCode.TxNotMined,
              OraklErrorCode.TxProcessingResponseError,
              OraklErrorCode.TxMissingResponseError
            ].includes(e.code)
          ) {
            throw e
          }

          logger.info(`Retrying transaction. Trial number: ${i}`)
        }
      }
    }
  }

  logger.debug('Reporter job built')
  return wrapper
}

// basically does the same job as reporter but checks if roundId is valid before reporting
export function dataFeedReporter(state: State, logger: Logger) {
  const reporterAggregateQueue = new Queue(REPORTER_AGGREGATOR_QUEUE_NAME, BULLMQ_CONNECTION)
  async function wrapper(job: Job) {
    const splittedJobId = job.id?.split('-')
    if (!splittedJobId || splittedJobId.length < 3) {
      throw new OraklError(
        OraklErrorCode.UnexpectedJobId,
        `unexpected jobId from dataFeedReporter: ${job.id}`
      )
    }
    const [roundId, oracleAddress, _] = splittedJobId

    if (!isRoundIdFresh(reporterAggregateQueue, oracleAddress, Number(roundId))) {
      throw new UnrecoverableError(`not reporting for low roundId: ${roundId}`)
    }
    await reporter(state, logger)(job)
  }
  return wrapper
}
