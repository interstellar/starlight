import { ChannelOp, WalletOp } from 'types/types'

export interface ApplicationState {
  config: ConfigState
  events: EventsState
  lifecycle: LifecycleState
  wallet: WalletState
  channels: ChannelsState
  flash: FlashState
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

export interface ChannelState {
  ChannelFeerate: number
  CounterpartyAddress: string
  EscrowAcct: string
  FinalityDelay: number
  FundingTime: string
  GuestAcct: string
  GuestAmount: number
  GuestRatchetAcct: string
  HostAcct: string
  HostAmount: number
  ID: string
  MaxRoundDuration: number
  Ops: ChannelOp[]
  PaymentTime: string
  PendingAmountReceived: number
  PendingAmountSent: number
  PrevState: string
  Role: string
  State: string
}

export interface FlashState {
  message: string
  showFlash: boolean
}
