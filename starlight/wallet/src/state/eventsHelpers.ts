import {
  AccountEvent,
  AccountReserveEvent,
  WalletOp,
  ChannelOp,
  ChannelEvent,
  ChannelTxEvent,
  ChannelMsgEvent,
  ChannelCmdEvent,
  ChannelTimeoutEvent,
  InputTx,
} from 'types/types'

const StrKey = require('stellar-base').StrKey

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

function isAccountReserveEvent(
  event: AccountEvent
): event is AccountReserveEvent {
  return event.InputCommand !== null
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
        // check if it's a funding transaction
        if (event.Channel.RoundNumber === 1) {
          return [
            {
              type: 'deposit',
              tx: event.InputTx,
              myDelta: isHost ? event.Channel.HostAmount : 0,
              theirDelta: isHost ? 0 : event.Channel.HostAmount,
              myBalance: 0,
              theirBalance: 0,
              isHost,
            },
          ]
        } else {
          const escrowAccountId = event.Channel.EscrowAcct
          const topUpAmount = getTopUpAmount(event.InputTx, escrowAccountId)
          const myDelta = isHost ? topUpAmount : 0
          const theirDelta = isHost ? 0 : topUpAmount
          return [
            {
              type: 'topUp',
              tx: event.InputTx,
              myDelta: isHost ? topUpAmount : 0,
              theirDelta: isHost ? 0 : topUpAmount,
              myBalance: myBalance - myDelta, // because we use the old balance
              theirBalance: theirBalance - theirDelta, // same
              isHost,
            },
          ]
        }
      }
      case 'Closed': {
        return [
          {
            type: 'withdrawal',
            tx: event.InputTx,
            myDelta: -1 * myBalance,
            theirDelta: -1 * theirBalance,
            myBalance,
            theirBalance,
            isHost,
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
                isHost,
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
                isHost,
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
            isHost,
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
                isHost,
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

function isChannelTxEvent(event: ChannelEvent): event is ChannelTxEvent {
  return (event as any).InputTx !== null
}

function isChannelMsgEvent(event: ChannelEvent): event is ChannelMsgEvent {
  return (event as any).InputMessage !== null
}

function isChannelCmdEvent(event: ChannelEvent): event is ChannelCmdEvent {
  return (event as any).InputCommand !== null
}

// TODO - this is currently unused
export function isChannelTimeoutEvent(
  event: ChannelEvent
): event is ChannelTimeoutEvent {
  return (event as any).InputLedgerTime !== '0001-01-01T00:00:00Z'
}

// TODO - this is currently unused
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
