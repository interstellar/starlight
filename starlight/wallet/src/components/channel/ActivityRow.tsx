import * as React from 'react'
import * as moment from 'moment'

import { DUSTYGRAY } from 'components/styled/Colors'
import { TableData } from 'components/styled/Table'
import { Timestamp } from 'components/styled/Timestamp'
import { ValueChange } from 'components/styled/ValueChange'
import { ChannelOp, ChannelState } from 'types'
import { stroopsToLumens } from 'lumens'

const activityTitle = (op: ChannelOp): string => {
  switch (op.type) {
    case 'deposit':
    case 'topUp':
      return 'Deposit'
    case 'incomingChannelPayment':
      return 'Receive'
    case 'outgoingChannelPayment':
      return 'Send'
    case 'withdrawal':
      return 'Withdraw'
    case 'paymentCompleted':
      throw new Error(
        `activityTitle shouldn't be called for ${op.type} op`
      )
  }
}

interface Props {
  state: ChannelState
  op: ChannelOp
  pending: boolean
  timestamp?: string
}

export class ActivityRow extends React.Component<Props, {}> {
  public constructor(props: any) {
    super(props)
  }

  public render() {
    const op = this.props.op
    if (op.type === 'paymentCompleted') {
      throw new Error(`ActivityRow should not be passed ${op.type} op`)
    }
    const time =
      op.type === 'deposit'
        ? moment(op.fundingTx.LedgerTime).fromNow()
        : this.props.timestamp
          ? moment(this.props.timestamp).fromNow()
          : ''
    const pendingParens =
      this.props.pending &&
      (op.type === 'incomingChannelPayment' ||
        op.type === 'outgoingChannelPayment')
    return (
      <tr>
        <TableData align="left">
          {activityTitle(op)} {pendingParens ? ' (pending)' : ''}{' '}
          <Timestamp>{time}</Timestamp>
        </TableData>
        <TableData align="right">
          <ValueChange value={op.myDelta} />
        </TableData>
        <TableData align="right" color={DUSTYGRAY}>
          {stroopsToLumens(op.myBalance + op.myDelta)} XLM
        </TableData>
        <TableData align="right">
          <ValueChange value={op.theirDelta} />
        </TableData>
        <TableData align="right" color={DUSTYGRAY}>
          {stroopsToLumens(op.theirBalance + op.theirDelta)} XLM
        </TableData>
      </tr>
    )
  }
}
