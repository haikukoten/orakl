import { describe, expect, beforeEach, test } from '@jest/globals'
import { listHandler, insertHandler, removeHandler } from '../src/cli/orakl-cli/src/listener'
import { openDb } from '../src/cli/orakl-cli/src/utils'
import { mkTmpFile } from '../src/utils'
import { TEST_MIGRATIONS_PATH } from '../src/settings'

describe('CLI Listener', function () {
  let DB
  const TMP_DB_FILE = mkTmpFile({ fileName: 'settings.test.sqlite' })
  const LISTENER = {
    chain: 'localhost',
    service: 'VRF',
    address: '0x0000000000000000000000000000000000000000',
    eventName: 'Event'
  }

  beforeEach(async () => {
    DB = await openDb({ dbFile: TMP_DB_FILE, migrate: true, migrationsPath: TEST_MIGRATIONS_PATH })
  })

  test('Should list all listeners', async function () {
    const listener = await listHandler(DB)({})
    expect(listener.length).toBeGreaterThan(0)
  })

  test('Should insert new listener', async function () {
    const listenerBefore = await listHandler(DB)({})
    await insertHandler(DB)(LISTENER)
    const listenerAfter = await listHandler(DB)({})
    expect(listenerAfter.length).toEqual(listenerBefore.length + 1)
  })

  test('Should delete listener based on id', async function () {
    const listenerBefore = await listHandler(DB)({})
    await removeHandler(DB)({ id: 1 })
    const listenerAfter = await listHandler(DB)({})
    expect(listenerAfter.length).toEqual(listenerBefore.length - 1)
  })
})