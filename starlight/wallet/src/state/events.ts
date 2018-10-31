import { EventsState } from 'types/schema'
import { CONFIG_INIT, CONFIG_EDIT } from 'state/config'
import { CHANNEL_UPDATE } from 'state/channels'
import { WALLET_UPDATE, ADD_WALLET_ACTIVITY } from 'state/wallet'
import { Update } from 'client/types'
import { initialClientState } from 'client/client'

// Actions
export const TX_SUCCESS = 'events/TX_SUCCESS'
export const TX_FAILED = 'events/TX_FAILED'
export const UPDATE_CLIENT_STATE = 'events/UPDATE_CLIENT_STATE'

// Reducer
const initialState: EventsState = {
  clientState: initialClientState,
}

const reducer = (state = initialState, action: any) => {
  switch (action.type) {
    case UPDATE_CLIENT_STATE: {
      return {
        ...state,
        clientState: action.clientState,
      }
    }
    default: {
      return state
    }
  }
}

// handler for events
const getHandler = (dispatch: any) => (update: Update) => {
  dispatch({
    type: UPDATE_CLIENT_STATE,
    clientState: update.ClientState,
  })

  switch (update.Type) {
    case 'initUpdate':
      dispatch({
        type: CONFIG_INIT,
        ...update.Config,
      })
      return dispatch({
        type: WALLET_UPDATE,
        Account: update.Account,
      })
    case 'configUpdate':
      return dispatch({ type: CONFIG_EDIT, ...update.Config })
    case 'accountUpdate': {
      return dispatch({
        type: WALLET_UPDATE,
        Account: update.Account,
      })
    }
    case 'walletActivityUpdate':
      return dispatch({
        type: ADD_WALLET_ACTIVITY,
        op: update.WalletOp,
        Account: update.Account,
      })
    case 'txSuccessUpdate':
      return dispatch({
        type: TX_SUCCESS,
        Seq: update.Tx.SeqNum,
      })
    case 'txFailureUpdate':
      return dispatch({
        type: TX_FAILED,
        Seq: update.Tx.SeqNum,
      })
    case 'channelUpdate': {
      dispatch({
        type: CHANNEL_UPDATE,
        channel: update.Channel,
        Ops: [],
      })
      return dispatch({
        type: WALLET_UPDATE,
        Account: update.Account,
      })
    }
    case 'channelActivityUpdate': {
      dispatch({
        type: CHANNEL_UPDATE,
        channel: update.Channel,
        Ops: [update.ChannelOp],
      })
      return dispatch({
        type: WALLET_UPDATE,
        Account: update.Account,
      })
    }
  }
}

export const events = {
  reducer,
  getHandler,
}
