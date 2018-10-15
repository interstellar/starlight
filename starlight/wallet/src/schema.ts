import { WalletOp, ChannelState } from 'types'

export interface ApplicationState {
  config: ConfigState
  events: EventsState
  lifecycle: LifecycleState
  wallet: WalletState
  channels: ChannelsState
}

export interface ConfigState {
  Username: string
  HorizonURL: string
}

export interface EventsState {
  From: number
  list: Array<{
    Type: string
    UpdateNum: number

    // TODO: include remaining event types
    Config: any
  }>
}

export interface LifecycleState {
  isConfigured: boolean
  isLoggedIn: boolean
}

export interface WalletState {
  ID: string
  Balance: number
  Ops: WalletOp[]
  Pending: { [s: string]: number }
}

export interface ChannelsState {
  [id: string]: ChannelState
}
