import { Starlightd } from 'lib/starlightd'
import { EventsState, ApplicationState } from 'schema'
import { CONFIG_INIT, CONFIG_EDIT } from 'state/config'
import { CHANNEL_UPDATE, getEscrowAccounts } from 'state/channels'
import { getWalletOp, getChannelOps } from 'state/eventsHelpers'
import { WALLET_UPDATE, ADD_WALLET_ACTIVITY } from 'state/wallet'
import { Event } from 'types'

const StrKey = require('stellar-base').StrKey

// Actions
export const EVENTS_RECEIVED = 'events/RECEIVED'
export const TX_SUCCESS = 'events/TX_SUCCESS'
export const TX_FAILED = 'events/TX_FAILED'

// Reducer
const initialState: EventsState = {
  From: 1,
  list: [],
}

const reducer = (state = initialState, action: any) => {
  switch (action.type) {
    case EVENTS_RECEIVED: {
      return {
        ...state,
        From: action.From,
      }
    }
    default: {
      return state
    }
  }
}

// Side effects
const fetch = async (dispatch: any, From: number) => {
  const response = await Starlightd.post(dispatch, '/api/updates', { From })

  if (response.ok && response.body.length >= 1) {
    dispatch({
      type: EVENTS_RECEIVED,
      From: From + response.body.length,
    })

    response.body.forEach((event: Event) => {
      switch (event.Type) {
        case 'init':
          dispatch({
            type: CONFIG_INIT,
            ...event.Config,
          })
          return dispatch({
            type: WALLET_UPDATE,
            Account: event.Account,
          })
        case 'config':
          return dispatch({ type: CONFIG_EDIT, ...event.Config })
        case 'account': {
          // use a thunk to get the state
          return dispatch((_: any, getState: () => ApplicationState) => {
            const op = getWalletOp(event)
            const state = getState()
            if (event.InputTx) {
              const sourceAccountEd25519 =
                event.InputTx.Env.Tx.SourceAccount.Ed25519
              const sourceAccount = StrKey.encodeEd25519PublicKey(
                sourceAccountEd25519
              )
              const escrowAccounts = getEscrowAccounts(state)
              const channelID = escrowAccounts[sourceAccount]
              if (channelID !== undefined) {
                // it's from a channel
                // handled elsewhere
                return
              }
            }
            return dispatch({
              type: ADD_WALLET_ACTIVITY,
              op,
              Account: event.Account,
            })
          })
        }
        case 'tx_success':
          return dispatch({
            type: TX_SUCCESS,
            Seq: event.InputTx.SeqNum,
          })
        case 'tx_failed':
          return dispatch({
            type: TX_FAILED,
            Seq: event.InputTx.SeqNum,
          })
        case 'channel': {
          const ops = getChannelOps(event)
          dispatch({
            type: CHANNEL_UPDATE,
            channel: event.Channel,
            Ops: ops,
          })
          return dispatch({
            type: WALLET_UPDATE,
            Account: event.Account,
          })
        }
      }
    })
  }

  return response.ok
}

export const events = {
  fetch,
  reducer,
}
