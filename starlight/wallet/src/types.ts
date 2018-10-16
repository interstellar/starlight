import { ChannelState } from 'schema'

export interface Credentials {
  Username: string
  Password: string
}

export type WalletOp =
  | CreateAccountOp
  | IncomingPaymentOp
  | OutgoingPaymentOp
  | AccountMergeOp

interface CreateAccountOp {
  type: 'createAccount'
  amount: number
  sourceAccount: string
  timestamp: string
}

interface IncomingPaymentOp {
  type: 'incomingPayment'
  amount: number
  sourceAccount: string
  timestamp: string
}

export interface OutgoingPaymentOp {
  type: 'outgoingPayment'
  amount: number
  recipient: string
  timestamp: string
  sequence: string
  pending: boolean
  failed: boolean
}

interface AccountMergeOp {
  type: 'accountMerge'
  amount: number
  sourceAccount: string
  timestamp: string
}

// XDR types
// skipping some fields we don't need

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

export type Event =
  | InitEvent
  | ConfigEvent
  | AccountEvent
  | TxSuccessEvent
  | TxFailedEvent
  | ChannelEvent

export interface InitEvent {
  Type: 'init'
  Account: {
    Balance: 0
    ID: string
  }
  Config: {
    Username: string
    HorizonURL: string
  }
  UpdateNum: number
  UpdateLedgerTime: string
}

export interface ConfigEvent {
  Type: 'config'
  Config: {
    Username: string
    HorizonURL: string
  }
  Account: Account
  UpdateNum: number
  UpdateLedgerTime: string
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

export interface TxSuccessEvent {
  Type: 'tx_success'
  Account: Account
  UpdateNum: number
  InputTx: InputTx
  UpdateLedgerTime: string
}

export interface TxFailedEvent {
  Type: 'tx_failed'
  UpdateNum: number
  InputTx: InputTx
  UpdateLedgerTime: string
}

export type AccountEvent = AccountPaymentEvent | AccountReserveEvent

export interface Account {
  Balance: number
  ID: string
}

export interface AccountPaymentEvent {
  Type: 'account'
  Account: Account
  InputTx: InputTx
  InputCommand: null
  OpIndex: number
  UpdateNum: number
  UpdateLedgerTime: string
}

export interface AccountReserveEvent {
  Type: 'account'
  Account: Account
  InputCommand: {
    Amount: number
    Recipient: string
    Time: string
    UserCommand: string
  }
  InputTx: null
  UpdateNum: number
  PendingSequence: string
  UpdateLedgerTime: string
}

export type ChannelEvent =
  | ChannelCmdEvent
  | ChannelTxEvent
  | ChannelMsgEvent
  | ChannelTimeoutEvent

export type ChannelCommand = CreateChannelCommand

export interface CreateChannelCommand {
  UserCommand: 'CreateChannel'
}

export type ChannelMsg = ChannelProposeMsg | ChannelAcceptMsg

export interface ChannelProposeMsg {
  ChannelID: string
  ChannelProposeMsg: any // TODO: use or remove?
}

export interface ChannelAcceptMsg {
  ChannelID: string
  ChannelAcceptMsg: any // TODO: use or remove?
}

export interface ChannelCmdEvent {
  Type: 'channel'
  Account: Account
  InputCommand: ChannelCommand
  InputTx: null
  InputMessage: null
  InputLedgerTime: '0001-01-01T00:00:00Z'
  UpdateNum: number
  Channel: ChannelState
  UpdateLedgerTime: string
}

export interface ChannelTxEvent {
  Type: 'channel'
  Account: Account
  InputTx: InputTx
  InputMessage: null
  InputCommand: null
  InputLedgerTime: '0001-01-01T00:00:00Z'
  UpdateNum: number
  PendingSequence: string
  Channel: ChannelState
  UpdateLedgerTime: string
}

export interface ChannelMsgEvent {
  Type: 'channel'
  Account: Account
  InputMessage: ChannelMsg
  InputTx: null
  InputCommand: null
  InputLedgerTime: '0001-01-01T00:00:00Z'
  UpdateNum: number
  PendingSequence: string
  Channel: ChannelState
  UpdateLedgerTime: string
}

export interface ChannelTimeoutEvent {
  Type: 'channel'
  Account: Account
  InputMessage: null
  InputTx: null
  InputCommand: null
  InputLedgerTime: string
  UpdateNum: number
  PendingSequence: string
  Channel: ChannelState
  UpdateLedgerTime: string
}

// channel operations

export type ChannelOp =
  | DepositOp
  | OutgoingChannelPaymentOp
  | IncomingChannelPaymentOp
  | ChannelPaymentCompletedOp
  | WithdrawalOp
  | TopUpOp

interface DepositOp {
  type: 'deposit'
  fundingTx: InputTx
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
}

interface TopUpOp {
  type: 'topUp'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  topUpTx: InputTx
}

export interface OutgoingChannelPaymentOp {
  type: 'outgoingChannelPayment'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
}

export interface IncomingChannelPaymentOp {
  type: 'incomingChannelPayment'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
}

export interface WithdrawalOp {
  type: 'withdrawal'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  withdrawalTx: InputTx
}

// channel activities
// you can turn a list of channel operations into a list of channel activities
// see state/channels.ts
export interface ChannelActivity {
  type: 'channelActivity'
  op:
    | DepositOp
    | OutgoingChannelPaymentOp
    | IncomingChannelPaymentOp
    | WithdrawalOp
    | TopUpOp
  timestamp?: string
  pending: boolean
  channelID: string
  counterparty: string
  isHost: boolean
}

export interface WalletActivity {
  type: 'walletActivity'
  op: WalletOp
  timestamp?: string
  pending: boolean
}

export type Activity = ChannelActivity | WalletActivity

// used to complete all payments
// until this op is broadcast, all payments are considered pending
export interface ChannelPaymentCompletedOp {
  type: 'paymentCompleted'
  timestamp: string
}
