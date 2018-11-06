import { Dispatch, Reducer } from 'redux'

import { ConfigState } from 'types/schema'
import { STATUS_UPDATE, LOGIN_SUCCESS } from 'state/lifecycle'
import { Starlightd } from 'lib/starlightd'

// Actions
export const CONFIG_INIT = 'settings/CONFIG_INIT'
export const CONFIG_EDIT = 'settings/CONFIG_EDIT'

// Reducer
const initialState: ConfigState = {
  Username: '',
  HorizonURL: '',
}

const reducer: Reducer<ConfigState> = (state = initialState, action) => {
  switch (action.type) {
    case CONFIG_INIT: {
      return {
        ...state,
        Username: action.Username,
        HorizonURL: action.HorizonURL,
      }
    }
    case CONFIG_EDIT: {
      return { ...state, HorizonURL: action.HorizonURL }
    }
    case LOGIN_SUCCESS: {
      return { ...state, Username: action.Username || '' }
    }
    default: {
      return state
    }
  }
}

export interface InitConfigParams {
  HorizonURL: string
  Password: string
  Username: string
}

// Side effects
const init = async (dispatch: Dispatch, params: InitConfigParams) => {
  const response = await Starlightd.client.configInit(params)

  const reducerParams = {
    Username: params.Username,
    HorizonURL: params.HorizonURL,
  }

  if (response.ok) {
    dispatch({ type: CONFIG_INIT, ...reducerParams })
    dispatch({
      type: STATUS_UPDATE,
      IsConfigured: true,
      IsLoggedIn: true,
    })
  } else {
    dispatch({
      type: STATUS_UPDATE,
      IsConfigured: false,
      IsLoggedIn: false,
    })
  }

  return response.ok
}

interface EditParams {
  HorizonURL?: string
  OldPassword?: string
  Password?: string
}

const edit = async (dispatch: Dispatch, params: EditParams) => {
  const response = await Starlightd.client.configEdit(params)

  if (response.ok) {
    dispatch({ type: CONFIG_EDIT, ...params })
  }

  return response.ok
}

export const config = {
  edit,
  init,
  reducer,
}
