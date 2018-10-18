import * as React from 'react'
import * as moment from 'moment'
import { connect } from 'react-redux'

import { ApplicationState } from 'types/schema'
import { Activity } from 'types/types'

import { ChannelActivityRow } from 'pages/wallet/ChannelActivityRow'
import { Table, TableHeaderRow, TableHeader } from 'pages/shared/Table'
import { WalletActivityRow } from 'pages/wallet/WalletActivityRow'

import { getWalletActivities } from 'state/wallet'
import { getChannelActivity } from 'state/channels'

interface Props {
  activity: Activity[]
}

export class WalletActivityTable extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    return (
      <Table>
        <thead>
          <TableHeaderRow>
            <TableHeader align="left">Type</TableHeader>
            <TableHeader align="left">Counterparty</TableHeader>
            <TableHeader align="right">Amount</TableHeader>
            <TableHeader align="right">Fee</TableHeader>
          </TableHeaderRow>
        </thead>
        <tbody>
          {this.props.activity
            .map((activity: Activity, i) => {
              return activity.type === 'channelActivity' ? (
                <ChannelActivityRow activity={activity} key={i} />
              ) : (
                <WalletActivityRow op={activity.op} key={i} />
              )
            })
            .reverse()}
        </tbody>
      </Table>
    )
  }
}

const mapStateToProps = (state: ApplicationState) => {
  // aggregate activity across channels and wallet
  const chans = Object.values(state.channels)
  const channelActivityArrays = chans.map(getChannelActivity)
  const walletActivity = getWalletActivities(state)
  const activities = ([] as Activity[]).concat(
    ...channelActivityArrays,
    walletActivity
  )
  activities.sort((a, b) => {
    if (a.timestamp === undefined) {
      return 1
    }
    if (b.timestamp === undefined) {
      return -1
    } else {
      return moment(a.timestamp).unix() - moment(b.timestamp).unix()
    }
  })

  return { activity: activities }
}
export const ConnectedWalletActivityTable = connect(
  mapStateToProps,
  {}
)(WalletActivityTable)
