import * as React from 'react'

import { validPublicKey } from 'helpers/account'
import { fromNowPast } from 'helpers/moment'

import { WalletOp } from 'types/types'

import { CopyableString } from 'pages/shared/CopyableString'
import { TableData } from 'pages/shared/Table'
import { Timestamp } from 'pages/shared/Timestamp'
import { ValueChange } from 'pages/shared/ValueChange'

interface Props {
  op: WalletOp
}

export class WalletActivityRow extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    switch (this.props.op.type) {
      case 'createAccount':
      case 'incomingPayment':
      case 'accountMerge': {
        // we can treat these the same
        return (
          <tr>
            <TableData align="left">
              Receive{' '}
              <Timestamp>{fromNowPast(this.props.op.timestamp)}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={this.props.op.sourceAccount}
                truncate={validPublicKey(this.props.op.sourceAccount)}
              />
            </TableData>
            <TableData align="right">
              <ValueChange value={this.props.op.amount} />
            </TableData>
            <TableData align="right">&mdash;</TableData>
          </tr>
        )
      }
      case 'outgoingPayment': {
        return (
          <tr>
            <TableData align="left">
              {this.props.op.pending
                ? 'Send (pending)'
                : this.props.op.failed
                  ? 'Send (failed)'
                  : 'Send'}{' '}
              <Timestamp>{fromNowPast(this.props.op.timestamp)}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={this.props.op.recipient}
                truncate={validPublicKey(this.props.op.recipient)}
              />
            </TableData>
            <TableData align="right">
              <ValueChange value={-1 * this.props.op.amount} />
            </TableData>
            <TableData align="right">
              <ValueChange value={-100} />
            </TableData>
          </tr>
        )
      }
    }
  }
}
