import { ChannelOp, WalletOp } from 'types/types'
import { ClientState } from 'starlight-sdk'

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
  clientState: ClientState
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
  AccountAddresses: { [s: string]: string }
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
  RoundNumber: number
  State: string
}

export interface FlashState {
  message: string
  color?: string
  showFlash: boolean
}
