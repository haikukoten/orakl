import { describe, expect, beforeEach, test } from '@jest/globals'
import { listHandler, insertHandler, removeHandler } from '../src/cli/orakl-cli/src/service'
import { openDb } from '../src/cli/orakl-cli/src/utils'
import { mkTmpFile } from '../src/utils'
import { TEST_MIGRATIONS_PATH } from '../src/settings'

describe('CLI Service', function () {
  let DB
  const TMP_DB_FILE = mkTmpFile({ fileName: 'settings.test.sqlite' })

  beforeEach(async () => {
    DB = await openDb({ dbFile: TMP_DB_FILE, migrate: true, migrationsPath: TEST_MIGRATIONS_PATH })
  })

  test('Should list service', async function () {
    const service = await listHandler(DB)()
    expect(service.length).toBeGreaterThan(0)
  })

  test('Should insert new service', async function () {
    const serviceBefore = await listHandler(DB)()
    await insertHandler(DB)({ name: 'Automation' })
    const serviceAfter = await listHandler(DB)()
    expect(serviceAfter.length).toEqual(serviceBefore.length + 1)
  })

  test('Should not allow to insert the same service more than once', async function () {
    await insertHandler(DB)({ name: 'Automation' })
    await expect(async () => {
      await insertHandler(DB)({ name: 'Automation' })
    }).rejects.toThrow()
  })

  test('Should delete service based on id', async function () {
    const serviceBefore = await listHandler(DB)()
    await removeHandler(DB)({ id: 1 })
    const serviceAfter = await listHandler(DB)()
    expect(serviceAfter.length).toEqual(serviceBefore.length - 1)
  })
})