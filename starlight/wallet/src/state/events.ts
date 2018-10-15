import { CONFIG_INIT, CONFIG_EDIT } from 'state/config'
import { EventsState, ApplicationState } from 'schema'
import { Starlightd } from 'lib/starlightd'
import { WALLET_UPDATE, ADD_WALLET_ACTIVITY } from 'state/wallet'
import { CHANNEL_UPDATE, getEscrowAccounts } from 'state/channels'
import {
  Event,
  AccountEvent,
  isAccountReserveEvent,
  WalletOp,
  ChannelOp,
  ChannelEvent,
  isChannelTxEvent,
  isChannelMsgEvent,
  isChannelCmdEvent,
  InputTx,
} from 'types'
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

// helper for parsing wallet events
export const getWalletOp = (event: AccountEvent): WalletOp => {
  if (isAccountReserveEvent(event)) {
    return {
      type: 'outgoingPayment',
      amount: event.InputCommand.Amount,
      recipient: event.InputCommand.Recipient,
      timestamp: event.InputCommand.Time,
      sequence: event.PendingSequence,
      pending: true,
      failed: false,
    }
  } else {
    const xdr = event.InputTx.Env
    const opIndex = event.OpIndex
    const tx = xdr.Tx
    const timestamp = event.InputTx.LedgerTime
    const op = tx.Operations[opIndex]
    const sourceAccountEd25519 = op.SourceAccount || tx.SourceAccount
    const sourceAccount = StrKey.encodeEd25519PublicKey(
      sourceAccountEd25519.Ed25519
    )
    const result = event.InputTx.Result.Result.Results[opIndex]
    const opBody = op.Body
    switch (opBody.Type) {
      case 0: {
        return {
          type: 'createAccount',
          amount: opBody.CreateAccountOp.StartingBalance,
          sourceAccount,
          timestamp,
        }
      }
      case 1: {
        // Payment op
        return {
          type: 'incomingPayment',
          amount: opBody.PaymentOp.Amount,
          sourceAccount,
          timestamp,
        }
      }
      case 8: {
        // AccountMerge op
        if (result.Tr.AccountMergeResult === null) {
          throw new Error('AccountMergeResult unexpectedly null')
        }
        return {
          type: 'accountMerge',
          amount: result.Tr.AccountMergeResult.SourceAccountBalance,
          sourceAccount,
          timestamp,
        }
      }
    }
    throw new Error('unexpectedly reached end of function')
  }
}

// helper for figuring out top-up amount
export const getTopUpAmount = (tx: InputTx, account: string) => {
  let total = 0
  for (const op of tx.Env.Tx.Operations) {
    if (op.Body.Type === 1) {
      const dest = StrKey.encodeEd25519PublicKey(
        op.Body.PaymentOp.Destination.Ed25519
      )
      if (dest === account) {
        total += op.Body.PaymentOp.Amount
      }
    }
  }
  return total
}

// helper for parsing channel events
export const getChannelOps = (event: ChannelEvent): ChannelOp[] => {
  const isHost = event.Channel.Role === 'Host'

  // note: these are the old balances, before the payment
  const myBalance = isHost
    ? event.Channel.HostAmount
    : event.Channel.GuestAmount
  const theirBalance = isHost
    ? event.Channel.GuestAmount
    : event.Channel.HostAmount

  if (isChannelTxEvent(event)) {
    switch (event.Channel.State) {
      case 'Open': {
        switch (event.Channel.PrevState) {
          case 'AwaitingFunding':
            return [
              {
                type: 'deposit',
                fundingTx: event.InputTx,
                myDelta: isHost ? event.Channel.HostAmount : 0,
                theirDelta: isHost ? 0 : event.Channel.HostAmount,
                myBalance: 0,
                theirBalance: 0,
              },
            ]
          case 'Open': {
            const escrowAccountId = event.Channel.EscrowAcct
            const topUpAmount = getTopUpAmount(event.InputTx, escrowAccountId)
            const myDelta = isHost ? topUpAmount : 0
            const theirDelta = isHost ? 0 : topUpAmount
            return [
              {
                type: 'topUp',
                topUpTx: event.InputTx,
                myDelta: isHost ? topUpAmount : 0,
                theirDelta: isHost ? 0 : topUpAmount,
                myBalance: myBalance - myDelta, // because we use the old balance
                theirBalance: theirBalance - theirDelta, // same
              },
            ]
          }
          default:
            return []
        }
      }
      case 'Closed': {
        return [
          {
            type: 'withdrawal',
            withdrawalTx: event.InputTx,
            myDelta: -1 * myBalance,
            theirDelta: -1 * theirBalance,
            myBalance,
            theirBalance,
          },
        ]
      }
    }
  } else if (isChannelMsgEvent(event)) {
    switch (event.Channel.State) {
      case 'Open': {
        switch (event.Channel.PrevState) {
          case 'PaymentAccepted':
          case 'PaymentProposed': {
            return [
              {
                type: 'paymentCompleted',
                timestamp: event.UpdateLedgerTime,
              },
            ]
          }
          default:
            return []
        }
      }
      case 'PaymentProposed': {
        switch (event.Channel.PrevState) {
          case 'PaymentProposed': {
            // it's a merge
            return [
              {
                type: 'incomingChannelPayment',
                myDelta: event.Channel.PendingAmountReceived,
                theirDelta: -1 * event.Channel.PendingAmountReceived,
                myBalance,
                theirBalance,
              },
            ]
          }
          default:
            return []
        }
      }
      case 'PaymentAccepted': {
        switch (event.Channel.PrevState) {
          case 'Open': {
            return [
              {
                type: 'incomingChannelPayment',
                myDelta: event.Channel.PendingAmountReceived,
                theirDelta: -1 * event.Channel.PendingAmountReceived,
                myBalance,
                theirBalance,
              },
            ]
          }
          case 'AwaitingMerge': {
            // we skip this because we already know about the payment
            return []
          }
          default:
            return []
        }
      }
      case 'AwaitingMerge': {
        // treat this as just an incoming payment proposal
        return [
          {
            type: 'incomingChannelPayment',
            myDelta: event.Channel.PendingAmountReceived,
            theirDelta: -1 * event.Channel.PendingAmountReceived,
            myBalance,
            theirBalance,
          },
        ]
      }
      default:
        return []
    }
  } else if (isChannelCmdEvent(event)) {
    switch (event.Channel.State) {
      case 'PaymentProposed': {
        switch (event.Channel.PrevState) {
          case 'Open': {
            return [
              {
                type: 'outgoingChannelPayment',
                myDelta: -1 * event.Channel.PendingAmountSent,
                theirDelta: event.Channel.PendingAmountSent,
                myBalance,
                theirBalance,
              },
            ]
          }
          default:
            return []
        }
      }
      case 'Open': {
        switch (event.Channel.PrevState) {
          case 'Open': {
            // top up command
            // TODO: mark this down as a pending top-up
          }
        }
      }
    }
  }
  return []
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
