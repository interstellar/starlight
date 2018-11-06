import configureStore from 'redux-mock-store'

import { CONFIG_INIT, CONFIG_EDIT } from 'state/config'
import { STATUS_UPDATE, LOGIN_SUCCESS } from 'state/lifecycle'
import { Starlightd as StarlightdImport } from 'lib/starlightd'
import { config } from 'state/config'
import { initialState } from 'state/testHelpers/initialState'

// hack to get around type safety when mocking
const Starlightd: any = StarlightdImport as any
Starlightd.client = {}

const mockStore = configureStore()

describe('reducer', () => {
  it('with CONFIG_INIT sets Username, HorizonURL', () => {
    const result = config.reducer(initialState.config, {
      type: CONFIG_INIT,
      Username: 'croaky',
      HorizonURL: 'https://horizon-testnet.stellar.org',
    })
    expect(result.Username).toBe('croaky')
    expect(result.HorizonURL).toBe('https://horizon-testnet.stellar.org')
  })

  it('with LOGIN_SUCCESS updates isLoggedIn', () => {
    const result = config.reducer(initialState.config, {
      type: LOGIN_SUCCESS,
      Username: 'croaky',
    })

    expect(result.Username).toBe('croaky')
  })
})

describe('init', () => {
  it('when ok, dispatch CONFIG_INIT, STATUS_UPDATE', async () => {
    const params = {
      Username: 'croaky',
      HorizonURL: 'slanket.com',
      Password: 'secret!',
    }
    const store = mockStore()

    Starlightd.client.configInit = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true, status: 200 }))

    await config.init(store.dispatch, params)

    expect(Starlightd.client.configInit).toHaveBeenCalledWith(params)
    expect(store.getActions()[0]).toEqual({
      type: CONFIG_INIT,
      Username: 'croaky',
      HorizonURL: 'slanket.com',
    })
    expect(store.getActions()[1]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: true,
      IsLoggedIn: true,
    })
  })

  it('when not ok, dispatch STATUS_UPDATE', async () => {
    const params = {
      Username: 'croaky',
      HorizonURL: '',
      Password: 'secret!',
    }
    const store = mockStore()
    Starlightd.client = {
      configInit: jest
        .fn()
        .mockImplementation(() => Promise.resolve({ ok: false })),
    }

    await config.init(store.dispatch, params)

    expect(Starlightd.client.configInit).toHaveBeenCalledWith(params)
    expect(store.getActions()[0]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: false,
      IsLoggedIn: false,
    })
  })
})

describe('edit', () => {
  it('when ok, dispatch CONFIG_EDIT', async () => {
    const store = mockStore()
    const params = { HorizonURL: 'boop' }
    Starlightd.client = {
      configEdit: jest
        .fn()
        .mockImplementation(() => Promise.resolve({ ok: true })),
    }

    await config.edit(store.dispatch, { HorizonURL: params.HorizonURL })

    expect(Starlightd.client.configEdit).toHaveBeenCalledWith({
      HorizonURL: params.HorizonURL,
    })
    expect(store.getActions()[0]).toEqual({ type: CONFIG_EDIT, ...params })
  })
})
