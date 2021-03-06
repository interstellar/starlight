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

export type ChannelOp =
  | DepositOp
  | OutgoingChannelPaymentOp
  | IncomingChannelPaymentOp
  | ChannelPaymentCompletedOp
  | WithdrawalOp
  | TopUpOp

// used to complete all payments
// until this op is broadcast, all payments are considered pending
export interface ChannelPaymentCompletedOp {
  type: 'paymentCompleted'
  timestamp: string
}

interface DepositOp {
  type: 'deposit'
  tx: InputTx
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  isHost: boolean
}

interface TopUpOp {
  type: 'topUp'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  tx: InputTx
  isHost: boolean
}

export interface OutgoingChannelPaymentOp {
  type: 'outgoingChannelPayment'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  isHost: boolean
}

export interface IncomingChannelPaymentOp {
  type: 'incomingChannelPayment'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  isHost: boolean
}

export interface WithdrawalOp {
  type: 'withdrawal'
  myDelta: number
  theirDelta: number
  myBalance: number
  theirBalance: number
  tx: InputTx
  isHost: boolean
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

export interface TxSuccessEvent {
  Type: 'tx_success'
  Account: Account
  UpdateNum: number
  InputTx: InputTx
  UpdateLedgerTime: string
}

export interface TxFailedEvent {
  Type: 'tx_failed'
  Account: Account
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
    Name: string
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
  Name: 'CreateChannel'
}

export type ChannelMsg = PaymentProposeMsg // only one needed for now

export interface PaymentProposeMsg {
  ChannelID: string
  PaymentProposeMsg: {
    PaymentAmount: number
    PaymentTime: string
    RoundNumber: number
  }
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
