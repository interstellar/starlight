import { Reducer, Dispatch } from 'redux'
import * as moment from 'moment'

import { ApplicationState, ChannelState, ChannelsState } from 'schema'
import { ChannelActivity } from 'types'
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
      if (oldChan === undefined) {
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

  if (response.ok) {
    await dispatch({
      type: CHANNEL_UPDATE,
      channel: response.body,
      Ops: [],
    })
  } else {
    console.log('error', response)
  }

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
  let pending = true
  let timestamp
  const activity: ChannelActivity[] = []
  for (let i = channel.Ops.length - 1; i >= 0; i--) {
    const channelOp = channel.Ops[i]
    if (channelOp.type === 'paymentCompleted') {
      pending = false
      timestamp = channelOp.timestamp
    } else if (channelOp.type === 'deposit') {
      activity.push({
        type: 'channelActivity',
        op: channelOp,
        timestamp: channelOp.fundingTx.LedgerTime,
        pending: false,
        channelID: channel.ID,
        counterparty: channel.CounterpartyAddress,
        isHost: channel.Role === 'Host',
      })
    } else if (channelOp.type === 'topUp') {
      activity.push({
        type: 'channelActivity',
        op: channelOp,
        timestamp: channelOp.topUpTx.LedgerTime,
        pending: false,
        channelID: channel.ID,
        counterparty: channel.CounterpartyAddress,
        isHost: channel.Role === 'Host',
      })
    } else if (channelOp.type === 'withdrawal') {
      activity.push({
        type: 'channelActivity',
        op: channelOp,
        timestamp: channelOp.withdrawalTx.LedgerTime,
        pending: false,
        channelID: channel.ID,
        counterparty: channel.CounterpartyAddress,
        isHost: channel.Role === 'Host',
      })
    } else {
      activity.push({
        type: 'channelActivity',
        op: channelOp,
        timestamp,
        pending,
        channelID: channel.ID,
        counterparty: channel.CounterpartyAddress,
        isHost: channel.Role === 'Host',
      })
    }
  }
  return activity
}

// Side effects
export const close = async (dispatch: Dispatch, id: string) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      UserCommand: 'CloseChannel',
    },
  })

  if (response.ok && response.body.length >= 1) {
    console.log(response.body)
  } else {
    console.log('error', response)
  }

  dispatch({
    type: 'TODO',
  })
}

export const channelPay = async (
  dispatch: Dispatch,
  id: string,
  amount: number
) => {
  const response = await Starlightd.post(dispatch, '/api/do-command', {
    ChannelID: id,
    Command: {
      UserCommand: 'ChannelPay',
      Amount: amount,
    },
  })

  if (response.ok && response.body.length >= 1) {
    console.log(response.body)
  } else {
    console.log('error', response)
  }

  dispatch({
    type: 'TODO',
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
      UserCommand: 'TopUp',
      Amount: amount,
    },
  })

  if (response.ok && response.body.length >= 1) {
    console.log(response.body)
  } else {
    console.log('error', response)
  }

  dispatch({
    type: 'TODO',
  })

  return response.ok
}

export const channels = {
  reducer,
}
