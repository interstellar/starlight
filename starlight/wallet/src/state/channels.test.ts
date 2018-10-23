import configureStore from 'redux-mock-store'

import { Starlightd } from 'lib/starlightd'
import { cancel } from 'state/channels'

const mockStore = configureStore()

describe('cancel', () => {
  it('sends a clean up request to the server', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await cancel(store.dispatch, '1')

    expect(Starlightd.post).toHaveBeenCalledWith(
      store.dispatch,
      '/api/do-command',
      {
        ChannelID: '1',
        Command: {
          UserCommand: 'CleanUp',
        },
      }
    )
  })
})
