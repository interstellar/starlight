import configureStore from 'redux-mock-store'

import { CONFIG_INIT, CONFIG_EDIT } from 'state/config'
import { STATUS_UPDATE, LOGIN_SUCCESS } from 'state/lifecycle'
import { Starlightd } from 'lib/starlightd'
import { config } from 'state/config'
import { initialState } from 'state/testHelpers/initialState'

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
  it('when ok (with custom url), dispatch CONFIG_INIT, STATUS_UPDATE', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await config.init(store.dispatch, {
      Username: 'croaky',
      HorizonURL: 'something.custom.com',
      Password: 'secret!',
      DemoServer: false,
    })

    expect(Starlightd.post).toHaveBeenCalledWith(
      store.dispatch,
      '/api/config-init',
      {
        Username: 'croaky',
        HorizonURL: 'something.custom.com',
        Password: 'secret!',
      }
    )
    expect(store.getActions()[0]).toEqual({
      type: CONFIG_INIT,
      Username: 'croaky',
      HorizonURL: 'something.custom.com',
    })
    expect(store.getActions()[1]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: true,
      IsLoggedIn: true,
    })
  })

  it('when ok (using demo server), dispatch CONFIG_INIT, STATUS_UPDATE', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await config.init(store.dispatch, {
      Username: 'croaky',
      HorizonURL: '',
      Password: 'secret!',
      DemoServer: true,
    })

    expect(Starlightd.post).toHaveBeenCalledWith(
      store.dispatch,
      '/api/config-init',
      {
        Username: 'croaky',
        HorizonURL: 'https://horizon-testnet.stellar.org',
        Password: 'secret!',
      }
    )
    expect(store.getActions()[0]).toEqual({
      type: CONFIG_INIT,
      Username: 'croaky',
      HorizonURL: 'https://horizon-testnet.stellar.org',
    })
    expect(store.getActions()[1]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: true,
      IsLoggedIn: true,
    })
  })

  it('when not ok, dispatch STATUS_UPDATE', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: false }))

    await config.init(store.dispatch, {
      Username: 'croaky',
      HorizonURL: '',
      Password: 'secret!',
      DemoServer: true,
    })

    expect(Starlightd.post).toHaveBeenCalledWith(
      store.dispatch,
      '/api/config-init',
      {
        Username: 'croaky',
        HorizonURL: 'https://horizon-testnet.stellar.org',
        Password: 'secret!',
      }
    )
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
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await config.edit(store.dispatch, { HorizonURL: params.HorizonURL })

    expect(Starlightd.post).toHaveBeenCalledWith(
      store.dispatch,
      '/api/config-edit',
      {
        HorizonURL: params.HorizonURL,
      }
    )
    expect(store.getActions()[0]).toEqual({ type: CONFIG_EDIT, ...params })
  })
})
