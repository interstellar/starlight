import configureStore from 'redux-mock-store'

import { initialState } from 'state/testHelpers/initialState'
import { lifecycle } from 'state/lifecycle'
import { Starlightd } from 'lib/starlightd'
import {
  LOGIN_SUCCESS,
  LOGIN_FAILURE,
  LOGOUT_SUCCESS,
  STATUS_UPDATE,
} from 'state/lifecycle'

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
    Starlightd.post = jest.fn().mockImplementation(() =>
      Promise.resolve({
        body: { IsConfigured: true },
        ok: true,
      })
    )

    await lifecycle.status(store.dispatch)

    expect(store.getActions()[0]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: true,
    })
    expect(Starlightd.post).toHaveBeenCalledWith(store.dispatch, '/api/status')
  })

  it('when not IsConfigured, dispatch STATUS_UPDATE', async () => {
    const store = mockStore()
    Starlightd.post = jest.fn().mockImplementation(() =>
      Promise.resolve({
        body: { IsConfigured: false },
        ok: true,
      })
    )

    await lifecycle.status(store.dispatch)

    expect(store.getActions()[0]).toEqual({
      type: STATUS_UPDATE,
      IsConfigured: false,
    })
    expect(Starlightd.post).toHaveBeenCalledWith(store.dispatch, '/api/status')
  })
})

describe('login', () => {
  it('when ok, dispatch LOGIN_SUCCESS', async () => {
    const store = mockStore()
    Starlightd.post = jest
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
    expect(Starlightd.post).toHaveBeenCalledWith(store.dispatch, '/api/login', {
      username: 'foo',
      password: 'bar',
    })
  })

  it('when not ok, dispatch LOGIN_FAILURE', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: false }))

    await lifecycle.login(store.dispatch, {
      Username: 'foo',
      Password: 'bar',
    })
    expect(store.getActions()[0]).toEqual({
      type: LOGIN_FAILURE,
    })
    expect(Starlightd.post).toHaveBeenCalledWith(store.dispatch, '/api/login', {
      username: 'foo',
      password: 'bar',
    })
  })
})

describe('logout', () => {
  it('when ok, dispatch LOGIN_SUCCESS', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ ok: true }))

    await lifecycle.logout(store.dispatch)

    expect(store.getActions()[0]).toEqual({
      type: LOGOUT_SUCCESS,
    })
    expect(Starlightd.post).toHaveBeenCalledWith(store.dispatch, '/api/logout')
  })

  it('when not ok, do not dispatch', async () => {
    const store = mockStore()
    Starlightd.post = jest
      .fn()
      .mockImplementation(() => Promise.resolve({ status: 404 }))

    await lifecycle.logout(store.dispatch)

    expect(store.getActions()[0]).toEqual(undefined)
    expect(Starlightd.post).toHaveBeenCalledWith(store.dispatch, '/api/logout')
  })
})
