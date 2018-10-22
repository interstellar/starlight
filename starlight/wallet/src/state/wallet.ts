import { Reducer, Dispatch } from 'redux'

import { WalletState, ApplicationState } from 'types/schema'
import { Starlightd } from 'lib/starlightd'
import { TX_SUCCESS, TX_FAILED } from 'state/events'
import { WalletActivity, OutgoingPaymentOp } from 'types/types'

// Actions
export const ADD_WALLET_ACTIVITY = 'wallet/ADD_WALLET_ACTIVITY'
export const WALLET_UPDATE = 'wallet/WALLET_UPDATE'

// Reducer
const initialState: WalletState = {
  ID: '',
  Balance: 0,
  Ops: [],
  Pending: {},
}

const reducer: Reducer<WalletState> = (
  state = initialState,
  action
): WalletState => {
  switch (action.type) {
    case WALLET_UPDATE: {
      return {
        ...state,
        ID: action.Account.ID,
        Balance: action.Account.Balance,
      }
    }
    case ADD_WALLET_ACTIVITY: {
      if (action.op.type === 'outgoingPayment') {
        const newActivityIndex = state.Ops.length
        const pendingSequence = action.op.sequence
        return {
          ...state,
          Balance: action.Account.Balance,
          Ops: [...state.Ops, action.op],
          Pending: {
            ...state.Pending,
            [pendingSequence]: newActivityIndex,
          },
        }
      }
      return {
        ...state,
        Balance: action.Account.Balance,
        Ops: [...state.Ops, action.op],
      }
    }
    case TX_SUCCESS: {
      const sequenceNumber = action.Seq
      const activityIndex = state.Pending[action.Seq]
      if (activityIndex === undefined) {
        return state
      }
      const newActivity = [...state.Ops]
      const payment = newActivity[activityIndex]
      const newPayment = {
        ...payment,
        pending: false,
      }
      newActivity[activityIndex] = newPayment
      return {
        ...state,
        Ops: newActivity,
        Pending: {
          ...state.Pending,
          [sequenceNumber]: undefined,
        },
      }
    }
    case TX_FAILED: {
      const sequenceNumber = action.Seq
      const activityIndex = state.Pending[action.Seq]
      if (activityIndex === undefined) {
        return state
      }
      const newActivity = [...state.Ops]
      const payment = newActivity[activityIndex] as OutgoingPaymentOp
      const newPayment: OutgoingPaymentOp = {
        ...payment,
        pending: false,
        failed: true,
      }
      newActivity[activityIndex] = newPayment
      return {
        ...state,
        Ops: newActivity,
        Pending: {
          ...state.Pending,
          [sequenceNumber]: undefined,
        },
      }
    }
    default: {
      return state
    }
  }
}

// selectors

export const getWalletState = (state: ApplicationState) => state.wallet

export const getWalletStroops = (state: ApplicationState) => {
  const walletState = getWalletState(state)
  return walletState.Balance
}

export const getWalletActivities = (
  state: ApplicationState
): WalletActivity[] => {
  const walletOps = state.wallet.Ops
  return walletOps.map(op => {
    const pending = op.type === 'outgoingPayment' && op.pending
    const walletActivity: WalletActivity = {
      type: 'walletActivity',
      op,
      timestamp: pending ? undefined : op.timestamp,
      pending,
    }
    return walletActivity
  })
}

// asynchronous

// Side effects
export const send = async (
  dispatch: Dispatch,
  recipient: string,
  amount: number
) => {
  const response = await Starlightd.post(dispatch, '/api/do-wallet-pay', {
    Dest: recipient,
    Amount: amount,
  })
  return response.ok
}

export const wallet = {
  reducer,
}
