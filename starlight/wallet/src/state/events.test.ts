import configureStore from 'redux-mock-store'

import { CONFIG_INIT } from 'state/config'
import { EVENTS_RECEIVED } from 'state/events'
import { Starlightd } from 'lib/starlightd'
import { events } from 'state/events'
import { initialState } from 'state/testHelpers/initialState'
import { WALLET_UPDATE } from 'state/wallet'

const mockStore = configureStore()

describe('reducer', () => {
  it('with EVENTS_RECEIVED returns state', () => {
    const result = events.reducer(initialState.events, {
      type: EVENTS_RECEIVED,
      From: 1,
    })

    expect(result.From).toEqual(1)
    expect(result.list).toEqual(initialState.events.list)
  })
})

describe('fetch', () => {
  it('when ok, dispatch EVENTS_RECEIVED, CONFIG_INIT', async () => {
    const store = mockStore()
    Starlightd.post = jest.fn().mockImplementation(() =>
      Promise.resolve({
        body: [
          {
            Type: 'init',
            UpdateNum: 1,
            Config: {
              Username: 'croaky',
              Password: '[redacted]',
            },
            ChannelInfo: null,
            Account: {
              ID: 'GDQEYK27FM4LZCV54D7XB75DR76BGJJYJEKNGREPAVARTYA27KHL6H32',
              Balance: 0,
            },
          },
        ],
        ok: true,
      })
    )

    await events.fetch(store.dispatch, 0)

    expect(Starlightd.post).toHaveBeenCalledWith(
      store.dispatch,
      '/api/updates',
      { From: 0 }
    )
    expect(store.getActions()[0]).toEqual({
      type: EVENTS_RECEIVED,
      From: 1,
    })
    expect(store.getActions()[1]).toEqual({
      type: CONFIG_INIT,
      Username: 'croaky',
      Password: '[redacted]',
    })
    expect(store.getActions()[2]).toEqual({
      type: WALLET_UPDATE,
      Account: {
        ID: 'GDQEYK27FM4LZCV54D7XB75DR76BGJJYJEKNGREPAVARTYA27KHL6H32',
        Balance: 0,
      },
    })
  })
})
