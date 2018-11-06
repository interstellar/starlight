import { ChannelOp, WalletOp } from 'types/types'

export interface ClientState {
  from: number
  addressesForCounterpartyAccount: { [s: string]: string }
  addressesForChannelAccount: { [s: string]: string }
  updateNumberForSequenceNumber: { [s: string]: number }
}

export type UpdateHandler = (event: Update) => void
export type ResponseHandler = (response: ClientResponse) => ClientResponse

// XDR types
// skipping some fields we don't need

export interface Account {
  Balance: number
  ID: string
}

export interface AccountID {
  Type: 0
  Ed25519: number[]
}

export interface CreateAccountOpBody {
  Type: 0
  CreateAccountOp: {
    Destination: AccountID
    StartingBalance: number
  }
}

export interface PaymentOpBody {
  Type: 1
  PaymentOp: {
    Destination: AccountID
    Amount: number
  }
}

export interface AccountMergeOpBody {
  Type: 8
  Destination: AccountID
}

export interface TxOperation {
  SourceAccount: AccountID | null
  Body: CreateAccountOpBody | PaymentOpBody | AccountMergeOpBody
}

export interface XdrEnvelope {
  Signatures: Array<{
    Hint: number[] // 4 bytes
    Signature: string
  }>
  Tx: {
    Fee: number
    Memo: any
    Operations: TxOperation[]
    SourceAccount: AccountID
  }
}

export interface Channel {
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
  PaymentTime: string
  PendingAmountReceived: number
  PendingAmountSent: number
  PrevState: string
  Role: string
  RoundNumber: number
  State: string
}

export interface XdrEnvelope {
  Signatures: Array<{
    Hint: number[] // 4 bytes
    Signature: string
  }>
  Tx: {
    Fee: number
    Memo: any
    Operations: TxOperation[]
    SourceAccount: AccountID
  }
}

export interface TxOperation {
  SourceAccount: AccountID | null
  Body: CreateAccountOpBody | PaymentOpBody | AccountMergeOpBody
}

export interface Result {
  Code: number
  Tr: {
    AccountMergeResult: null | { Code: 0; SourceAccountBalance: number }
  }
}

export interface InputTx {
  Env: XdrEnvelope
  LedgerNum: number
  LedgerTime: string
  PT: string
  Result: {
    FeeCharged: number
    Result: {
      Code: number
      Results: Result[]
    }
    // TBD: rest of this
  }
  SeqNum: string
}

export interface GenericUpdate {
  Account: Account
  UpdateNum: number
  UpdateLedgerTime: string
  ClientState: ClientState
}

export type Update =
  | InitUpdate
  | ConfigUpdate
  | AccountUpdate
  | WalletActivityUpdate
  | ChannelUpdate
  | ChannelActivityUpdate
  | TxUpdate

export interface InitUpdate extends GenericUpdate {
  Type: 'initUpdate'
  Config: {
    Username: string
    HorizonURL: string
  }
}

export interface ConfigUpdate extends GenericUpdate {
  Type: 'configUpdate'
  Config: {
    Username: string
    HorizonURL: string
  }
}

interface AccountUpdate extends GenericUpdate {
  Type: 'accountUpdate'
}

interface WalletActivityUpdate extends GenericUpdate {
  Type: 'walletActivityUpdate'
  WalletOp: WalletOp
}

interface ChannelUpdate extends GenericUpdate {
  Type: 'channelUpdate'
  Channel: Channel
}

interface ChannelActivityUpdate extends GenericUpdate {
  Type: 'channelActivityUpdate'
  Channel: Channel
  ChannelOp: ChannelOp
}

interface TxUpdate extends GenericUpdate {
  Type: 'txSuccessUpdate' | 'txFailureUpdate'
  Tx: InputTx
}

export interface WalletActivity {
  type: 'walletActivity'
  delta: number // positive or negative
  counterparty: string // either sender or recipient
}

export interface ClientResponse {
  body: any
  ok: boolean
  status?: number
  error?: Error
}
