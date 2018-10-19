import { Dispatch, Reducer } from 'redux'

import { timer } from 'helpers/timer'
import { FlashState } from 'types/schema'

// Actions
export const SET_FLASH = 'flash/SET'
export const CLEAR_FLASH = 'flash/CLEAR'

// Reducer
export const initialState: FlashState = {
  message: '',
  showFlash: false,
}

const reducer: Reducer<FlashState> = (state = initialState, action) => {
  switch (action.type) {
    case SET_FLASH: {
      return { ...state, message: action.message, showFlash: true }
    }
    case CLEAR_FLASH: {
      return { ...state, message: '', showFlash: false }
    }
    default: {
      return state
    }
  }
}

// side effects
const set = async (dispatch: Dispatch, message: string) => {
  dispatch({ type: SET_FLASH, message })

  timer(() => dispatch({ type: CLEAR_FLASH }), 3000)
}

export const flash = {
  set,
  reducer,
}
