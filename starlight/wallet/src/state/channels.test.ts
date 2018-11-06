import configureStore from 'redux-mock-store'

import { Starlightd as StarlightdImport } from 'lib/starlightd'
import { cancel, forceClose } from 'state/channels'

// hack to get around type safety when mocking
const Starlightd: any = StarlightdImport as any
Starlightd.client = {}

const mockStore = configureStore()

describe('cancel', () => {
  it('sends a clean up request to the server', async () => {
    const store = mockStore()
    Starlightd.client.cancel = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await cancel(store.dispatch, '1')

    expect(Starlightd.client.cancel).toHaveBeenCalledWith('1')
  })
})

describe('force close', () => {
  it('sends a force close request to the server', async () => {
    const store = mockStore()
    Starlightd.client.forceClose = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await forceClose(store.dispatch, '1')

    expect(Starlightd.client.forceClose).toHaveBeenCalledWith('1')
  })
})
