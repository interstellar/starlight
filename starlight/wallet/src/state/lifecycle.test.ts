import configureStore from 'redux-mock-store'

import { initialState } from 'state/testHelpers/initialState'
import {
  lifecycle,
  LOGIN_SUCCESS,
  LOGIN_FAILURE,
  LOGOUT_SUCCESS,
  STATUS_UPDATE,
} from 'state/lifecycle'
import { Starlightd as StarlightdImport } from 'lib/starlightd'

// hack to get around type safety when mocking
const Starlightd: any = StarlightdImport as any
Starlightd.client = {}

const mockStore = configureStore()

describe('reducer', () => {
  it('with STATUS_UPDATE updates isConfigured, isLoggedIn', () => {
    let result = lifecycle.reducer(initialState.lifecycle, {
      type: STATUS_UPDATE,
      IsConfigured: true,
      IsLoggedIn: true,
    })

    expect(result.isConfigured).toBe(true)

    result = lifecycle.reducer(initialState.lifecycle, {
      type: STATUS_UPDATE,
      IsConfigured: false,
      IsLoggedIn: false,
    })

    expect(result.isConfigured).toBe(false)
  })

  it('with LOGIN_SUCCESS updates isLoggedIn', () => {
    const result = lifecycle.reducer(initialState.lifecycle, {
      type: LOGIN_SUCCESS,
    })
    expect(result.isLoggedIn).toBe(true)
  })

  it('with LOGIN_FAILURE updates isLoggedIn', () => {
    const result = lifecycle.reducer(initialState.lifecycle, {
      type: LOGIN_FAILURE,
    })

    expect(result.isLoggedIn).toBe(false)
  })

  it('with LOGOUT_SUCCESS updates isLoggedIn', () => {
    const result = lifecycle.reducer(initialState.lifecycle, {
      type: LOGOUT_SUCCESS,
    })

    expect(result.isLoggedIn).toBe(false)
  })
})

describe('status', () => {
  it('when IsConfigured, dispatch STATUS_UPDATE', async () => {
    const store = mockStore()
    Starlightd.client.getStatus = jest.fn().mockImplementation(() =>
      Promise.resolve({
        body: { IsConfigured: true, IsLoggedIn: true },
        ok: true,
        status: 200,
      })
    )

    await lifecycle.status(store.dispatch)

    expect(store.getActions()[0]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: true,
      IsLoggedIn: true,
    })
    expect(Starlightd.client.getStatus).toHaveBeenCalled()
  })

  it('when not IsConfigured, dispatch STATUS_UPDATE', async () => {
    const store = mockStore()
    Starlightd.client.getStatus = jest.fn().mockImplementation(() =>
      Promise.resolve({
        body: { IsConfigured: false, IsLoggedIn: false },
        ok: true,
        status: 200,
      })
    )

    await lifecycle.status(store.dispatch)

    expect(store.getActions()[0]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: false,
      IsLoggedIn: false,
    })
    expect(Starlightd.client.getStatus).toHaveBeenCalled()
  })
})

describe('login', () => {
  it('when ok, dispatch LOGIN_SUCCESS', async () => {
    const store = mockStore()
    Starlightd.client.login = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await lifecycle.login(store.dispatch, {
      Username: 'foo',
      Password: 'bar',
    })
    expect(store.getActions()[0]).toEqual({
      type: LOGIN_SUCCESS,
      Username: 'foo',
    })
    expect(Starlightd.client.login).toHaveBeenCalledWith('foo', 'bar')
  })

  it('when not ok, dispatch LOGIN_FAILURE', async () => {
    const store = mockStore()
    Starlightd.client.login = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: false }))

    await lifecycle.login(store.dispatch, {
      Username: 'foo',
      Password: 'bar',
    })
    expect(store.getActions()[0]).toEqual({
      type: LOGIN_FAILURE,
    })
    expect(Starlightd.client.login).toHaveBeenCalledWith('foo', 'bar')
  })
})

describe('logout', () => {
  it('when ok, dispatch LOGOUT_SUCCESS', async () => {
    const store = mockStore()
    Starlightd.client.logout = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    Starlightd.client.clearState = jest.fn()

    await lifecycle.logout(store.dispatch)

    expect(store.getActions()[0]).toEqual({
      type: LOGOUT_SUCCESS,
    })
    expect(Starlightd.client.logout).toHaveBeenCalled()
    expect(Starlightd.client.clearState).toHaveBeenCalled()
  })

  it('when not ok, do not dispatch', async () => {
    const store = mockStore()
    Starlightd.client.logout = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: false, status: 404 }))

    await lifecycle.logout(store.dispatch)

    expect(store.getActions()[0]).toEqual(undefined)
    expect(Starlightd.client.login).toHaveBeenCalled()
  })
})
