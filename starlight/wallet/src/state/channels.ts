import { Reducer, Dispatch } from 'redux'
import * as moment from 'moment'

import { ApplicationState, ChannelState, ChannelsState } from 'types/schema'
import { ChannelActivity } from 'types/types'
import { Starlightd } from 'lib/starlightd'

// Actions
export const CHANNEL_UPDATE = 'channels/CHANNEL_UPDATE'
export const CHANNEL_CLOSE = 'channels/CHANNEL_CLOSE'

// helpers
export const getMyBalance = (channel: ChannelState) => {
  if (channel.State === 'Closed' || channel.State === '') {
    return 0
  }
  return channel.Role === 'Guest' ? channel.GuestAmount : channel.HostAmount
}

export const getTheirBalance = (channel: ChannelState) => {
  if (channel.State === 'Closed' || channel.State === '') {
    return 0
  }
  return channel.Role === 'Guest' ? channel.HostAmount : channel.GuestAmount
}

export const getMyAccount = (channel: ChannelState) => {
  return channel.Role === 'Guest' ? channel.GuestAcct : channel.HostAcct
}

export const getTheirAccount = (channel: ChannelState) => {
  return channel.Role === 'Guest' ? channel.HostAcct : channel.GuestAcct
}

export const getWithdrawalTime = (channel: ChannelState) => {
  const durationInNanoseconds =
    2 * channel.FinalityDelay + channel.MaxRoundDuration
  const durationInSeconds = durationInNanoseconds / 1000000000

  const withdrawalTime = moment(channel.PaymentTime)
    .add(durationInSeconds, 'seconds')
    .format('LLL')

  return `${withdrawalTime} (${moment(withdrawalTime).fromNow()})`
}

// Reducer
const initialState: ChannelsState = {}

const reducer: Reducer<ChannelsState> = (
  state = initialState,
  action: any
): ChannelsState => {
  switch (action.type) {
    case CHANNEL_UPDATE:
      const channel = action.channel as ChannelState
      const channelID = channel.CounterpartyAddress
      const oldChan = state[channelID]
      if (oldChan === undefined || oldChan.State === 'Closed') {
        return {
          ...state,
          [channelID]: {
            ...Object.assign(action.channel, { Ops: [...action.Ops] }),
          },
        }
      } else {
        return {
          ...state,
          [channelID]: {
            ...Object.assign(action.channel, {
              Ops: [...oldChan.Ops, ...action.Ops],
            }),
          },
        }
      }
    default:
      return state
  }
}

// Side effects
export const createChannel = async (
  dispatch: Dispatch,
  GuestAddr: string,
  HostAmount: number
) => {
  const response = await Starlightd.post(dispatch, '/api/do-create-channel', {
    GuestAddr,
    HostAmount,
  })

  if (response.ok && response.body !== '') {
    await dispatch({
      type: CHANNEL_UPDATE,
      channel: response.body,
      Ops: [],
    })
    return true
  } else {
    console.log('error', response)
    return false
  }
}

export const cancel = async (dispatch: Dispatch, id: string) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      Name: 'CleanUp',
    },
  })
  return response.ok
}

export const close = async (dispatch: Dispatch, id: string) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      Name: 'CloseChannel',
    },
  })
  return response.ok
}

export const forceClose = async (dispatch: Dispatch, id: string) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      Name: 'ForceClose',
    },
  })
  return response.ok
}

export const channelPay = async (
  dispatch: Dispatch,
  id: string,
  amount: number
) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      Name: 'ChannelPay',
      Amount: amount,
    },
  })
  return response.ok
}

export const deposit = async (
  dispatch: Dispatch,
  id: string,
  amount: number
) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      Name: 'TopUp',
      Amount: amount,
    },
  })
  return response.ok
}

// selectors
export const getChannels = (state: ApplicationState) => {
  const chans = Object.values(state.channels)
  chans.sort(
    (a, b) => moment(a.FundingTime).unix() - moment(b.FundingTime).unix()
  )
  return chans
}

export const getCounterpartyAccounts = (state: ApplicationState) => {
  const chans = Object.values(state.channels)
  const accounts: { [id: string]: string } = {}

  // accounts to channel IDs
  chans.forEach((chan: ChannelState) => {
    accounts[getTheirAccount(chan)] = chan.CounterpartyAddress
  })
  return accounts
}

export const getEscrowAccounts = (state: ApplicationState) => {
  const chans = Object.values(state.channels)
  const accounts: { [id: string]: string } = {}

  // escrow accounts to channel IDs
  chans.forEach((chan: ChannelState) => {
    accounts[chan.EscrowAcct] = chan.CounterpartyAddress
  })
  return accounts
}

export const getNumberOfOpenHostChannels = (state: ApplicationState) => {
  const chans = Object.values(state.channels)
  return chans
    .filter(chan => chan.Role === 'Host')
    .filter(chan => chan.State !== 'Closed').length
}

export const getTotalChannelBalance = (state: ApplicationState) => {
  const chans = Object.values(state.channels)
  return chans
    .map(chan => getMyBalance(chan))
    .reduce((prev, cur) => prev + cur, 0)
}

export const getTotalChannelCounterpartyBalance = (state: ApplicationState) => {
  const chans = Object.values(state.channels)
  return chans
    .map(chan => getTheirBalance(chan))
    .reduce((prev, cur) => prev + cur, 0)
}

export const getChannelActivity = (channel: ChannelState) => {
  let pendingPayments = true
  let timestamp

  const activities: ChannelActivity[] = []

  for (let i = channel.Ops.length - 1; i >= 0; i--) {
    const op = channel.Ops[i]

    if (op.type === 'paymentCompleted') {
      // all payments ('outgoingChannelPayment' & 'incomingChannelPayment')
      // are considered pending until / unless a 'paymentCompleted' op occurs
      pendingPayments = false
      timestamp = op.timestamp
    } else {
      if (
        op.type === 'deposit' ||
        op.type === 'topUp' ||
        op.type === 'withdrawal'
      ) {
        activities.push({
          type: 'channelActivity',
          op,
          timestamp: op.tx.LedgerTime,
          pending: false,
          channelID: channel.ID,
          counterparty: channel.CounterpartyAddress,
        })
      } else if (
        op.type === 'outgoingChannelPayment' ||
        op.type === 'incomingChannelPayment'
      ) {
        activities.push({
          type: 'channelActivity',
          op,
          timestamp,
          pending: pendingPayments,
          channelID: channel.ID,
          counterparty: channel.CounterpartyAddress,
        })
      }
    }
  }
  return activities
}

export const channels = {
  reducer,
}
