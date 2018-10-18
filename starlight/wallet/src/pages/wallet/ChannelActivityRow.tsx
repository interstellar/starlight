import * as React from 'react'
import * as moment from 'moment'

import { ChannelActivity } from 'types'

import { CopyableString } from 'pages/shared/CopyableString'
import { TableData } from 'pages/shared/Table'
import { Timestamp } from 'pages/shared/Timestamp'
import { ValueChange } from 'pages/shared/ValueChange'

import { stroopsToLumens } from 'helpers/lumens'

const StrKey = require('stellar-base').StrKey

interface Props {
  activity: ChannelActivity
}

export class ChannelActivityRow extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    const timestamp =
      this.props.activity.timestamp !== undefined
        ? moment(this.props.activity.timestamp).fromNow()
        : ''
    const activity = this.props.activity
    const op = this.props.activity.op
    switch (op.type) {
      case 'deposit':
      case 'topUp': {
        if (!activity.isHost) {
          return <tr />
        }
        const fee = op.tx.Env.Tx.Fee
        return (
          <tr>
            <TableData align="left">
              Channel deposit <Timestamp>{timestamp}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={activity.counterparty}
                truncate={StrKey.isValidEd25519PublicKey(activity.counterparty)}
              />
            </TableData>
            <TableData align="right">
              {stroopsToLumens(Math.abs(op.myDelta))} XLM
            </TableData>
            <TableData align="right">
              <ValueChange value={-1 * fee} />
            </TableData>
          </tr>
        )
      }
      case 'withdrawal': {
        if (op.myDelta === 0) {
          return <tr />
        }
        const fee = this.props.activity.isHost ? -1 * op.tx.Env.Tx.Fee : 0
        return (
          <tr>
            <TableData align="left">
              Channel withdrawal <Timestamp>{timestamp}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={activity.counterparty}
                truncate={StrKey.isValidEd25519PublicKey(activity.counterparty)}
              />
            </TableData>
            <TableData align="right">
              {stroopsToLumens(Math.abs(op.myDelta))} XLM
            </TableData>
            <TableData align="right">
              <ValueChange value={fee} />
            </TableData>
          </tr>
        )
      }
      case 'outgoingChannelPayment':
      case 'incomingChannelPayment': {
        return (
          <tr>
            <TableData align="left">
              {op.type === 'outgoingChannelPayment'
                ? 'Send via channel'
                : 'Receive via channel'}
              {activity.pending ? ' (pending)' : ''}{' '}
              <Timestamp>{timestamp}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={activity.counterparty}
                truncate={StrKey.isValidEd25519PublicKey(activity.counterparty)}
              />
            </TableData>
            <TableData align="right">
              <ValueChange value={op.myDelta} />
            </TableData>
            <TableData align="right">&mdash;</TableData>
          </tr>
        )
      }
      case 'withdrawal': {
        return (
          <tr>
            <TableData align="left">
              Withdraw <Timestamp>{timestamp}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={activity.counterparty}
                truncate={StrKey.isValidEd25519PublicKey(activity.counterparty)}
              />
            </TableData>
            <TableData align="right">
              <ValueChange value={-1 * op.myDelta} />
            </TableData>
            <TableData align="right">
              <ValueChange value={-1 * op.tx.Env.Tx.Fee} />
            </TableData>
          </tr>
        )
      }
    }
  }
}
