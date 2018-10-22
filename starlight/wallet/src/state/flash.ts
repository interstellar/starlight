import { Dispatch, Reducer } from 'redux'

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
      return {
        ...state,
        message: action.message,
        color: action.color,
        showFlash: true,
      }
    }
    case CLEAR_FLASH: {
      return { ...state, message: '', color: '', showFlash: false }
    }
    default: {
      return state
    }
  }
}

// side effects
const set = async (dispatch: Dispatch, message: string, color?: string) => {
  dispatch({ type: SET_FLASH, message, color })

  setTimeout(() => {
    dispatch({ type: CLEAR_FLASH })
  }, 3000)
}

export const flash = {
  set,
  reducer,
}
