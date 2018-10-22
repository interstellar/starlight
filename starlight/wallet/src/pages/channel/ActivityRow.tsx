import * as React from 'react'
import * as moment from 'moment'
import styled from 'styled-components'

import { DUSTYGRAY } from 'pages/shared/Colors'
import { TableData } from 'pages/shared/Table'
import { Timestamp } from 'pages/shared/Timestamp'
import { ValueChange } from 'pages/shared/ValueChange'
import { ChannelOp } from 'types/types'
import { ChannelState } from 'types/schema'
import { formatAmount, stroopsToLumens } from 'helpers/lumens'

const Row = styled.tr<{ pending: boolean }>`
  ${props => (props.pending && 'opacity: 0.5; font-style: italic;') }
`

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
      throw new Error(`activityTitle shouldn't be called for ${op.type} op`)
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
        ? moment(op.tx.LedgerTime).fromNow()
        : this.props.timestamp
          ? moment(this.props.timestamp).fromNow()
          : ''
    const pendingPayment =
      this.props.pending &&
      (op.type === 'incomingChannelPayment' ||
        op.type === 'outgoingChannelPayment')
    return (
      <Row pending={pendingPayment}>
        <TableData align="left">
          {activityTitle(op)} {pendingPayment ? ' (pending)' : ''}{' '}
          <Timestamp>{time}</Timestamp>
        </TableData>
        <TableData align="right">
          <ValueChange value={op.myDelta} />
        </TableData>
        <TableData align="right" color={DUSTYGRAY}>
          {formatAmount(stroopsToLumens(op.myBalance + op.myDelta))} XLM
        </TableData>
        <TableData align="right">
          <ValueChange value={op.theirDelta} />
        </TableData>
        <TableData align="right" color={DUSTYGRAY}>
          {formatAmount(stroopsToLumens(op.theirBalance + op.theirDelta))} XLM
        </TableData>
      </Row>
    )
  }
}
