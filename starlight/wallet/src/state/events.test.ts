import configureStore from 'redux-mock-store'

import { initialClientState, Client } from 'client/client'
import { CONFIG_INIT } from 'state/config'
import { events, UPDATE_CLIENT_STATE } from 'state/events'
import { initialState } from 'state/testHelpers/initialState'
import { WALLET_UPDATE } from 'state/wallet'

const mockStore = configureStore()

describe('reducer', () => {
  it('with UPDATE_CLIENT_STATE returns state', () => {
    const result = events.reducer(initialState.events, {
      type: UPDATE_CLIENT_STATE,
      clientState: {
        ...initialClientState,
        from: 1,
      },
    })

    expect(result.clientState.from).toEqual(1)
  })
})

describe('fetch', () => {
  it('when ok, dispatch UPDATE_CLIENT_STATE, CONFIG_INIT', async () => {
    const store = mockStore()
    const response = {
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
      loggedIn: true,
    }

    const client = new Client(initialClientState, () => undefined)
    client.handler = events.getHandler(store.dispatch)
    await client.handleResponse(response)

    expect(store.getActions()[0]).toEqual({
      type: UPDATE_CLIENT_STATE,
      clientState: {
        ...initialClientState,
        from: 2,
      },
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
