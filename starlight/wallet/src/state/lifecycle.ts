import { Dispatch, Reducer } from 'redux'

import { Credentials } from 'types/types'
import { LifecycleState } from 'types/schema'
import { Starlightd } from 'lib/starlightd'

// Actions
export const STATUS_UPDATE = 'lifecycle/STATUS_UPDATE'
export const LOGIN_SUCCESS = 'lifecycle/LOGIN_SUCCESS'
export const LOGIN_FAILURE = 'lifecycle/LOGIN_FAILURE'
export const LOGOUT_SUCCESS = 'lifecycle/LOGOUT_SUCCESS'

// Reducer
const initialState: LifecycleState = {
  isLoggedIn: false,
  isConfigured: false,
}

const reducer: Reducer<LifecycleState> = (state = initialState, action) => {
  switch (action.type) {
    case STATUS_UPDATE:
      return {
        ...state,
        isConfigured: action.IsConfigured,
        isLoggedIn: action.IsLoggedIn,
      }
    case LOGIN_SUCCESS: {
      return { ...state, isLoggedIn: true }
    }
    case LOGOUT_SUCCESS:
    case LOGIN_FAILURE: {
      return { ...state, isLoggedIn: false }
    }
    default: {
      return state
    }
  }
}

// Side effects
const status = async (dispatch: Dispatch) => {
  const response = await Starlightd.post(dispatch, '/api/status')

  if (response.ok) {
    dispatch({ type: STATUS_UPDATE, ...response.body })
  }
}

const login = async (dispatch: Dispatch, params: Credentials) => {
  const response = await Starlightd.post(dispatch, '/api/login', {
    username: params.Username,
    password: params.Password,
  })

  if (response.ok) {
    dispatch({ type: LOGIN_SUCCESS, Username: params.Username })
  } else {
    dispatch({ type: LOGIN_FAILURE })
  }

  return response.ok
}

const logout = async (dispatch: Dispatch) => {
  const response = await Starlightd.post(dispatch, '/api/logout')

  if (response.ok) {
    dispatch({ type: LOGOUT_SUCCESS })
  }

  return response.ok
}

export const lifecycle = {
  reducer,
  status,
  login,
  logout,
}
