import { combineReducers } from 'redux'

import { ApplicationState } from 'types/schema'
import { config } from 'state/config'
import { events } from 'state/events'
import { lifecycle, LOGOUT_SUCCESS } from 'state/lifecycle'
import { wallet } from 'state/wallet'
import { channels } from 'state/channels'
import { flash } from 'state/flash'

const appReducer = combineReducers<ApplicationState>({
  config: config.reducer,
  events: events.reducer,
  lifecycle: lifecycle.reducer,
  wallet: wallet.reducer,
  channels: channels.reducer,
  flash: flash.reducer,
})

export const rootReducer = (state: any, action: any) => {
  if (action.IsConfigured === false) {
    state = {
      lifecycle: state.lifecycle,
    }
  } else if (action.type === LOGOUT_SUCCESS || action.IsLoggedIn === false) {
    state = {
      config: state.config,
      lifecycle: state.lifecycle,
    }
  }

  return appReducer(state, action)
}
