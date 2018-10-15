import * as React from 'react'
import * as moment from 'moment'

import { WalletOp } from 'types'

import { CopyableString } from 'components/styled/CopyableString'
import { TableData } from 'components/styled/Table'
import { Timestamp } from 'components/styled/Timestamp'
import { ValueChange } from 'components/styled/ValueChange'

const StrKey = require('stellar-base').StrKey

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
              <Timestamp>{moment(this.props.op.timestamp).fromNow()}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={this.props.op.sourceAccount}
                truncate={StrKey.isValidEd25519PublicKey(this.props.op.sourceAccount)}
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
              <Timestamp>{moment(this.props.op.timestamp).fromNow()}</Timestamp>
            </TableData>
            <TableData align="left">
              <CopyableString
                id={this.props.op.recipient}
                truncate={StrKey.isValidEd25519PublicKey(this.props.op.recipient)}
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
