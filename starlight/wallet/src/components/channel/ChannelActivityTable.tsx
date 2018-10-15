import * as React from 'react'

import { ChannelState } from 'types'

import { ActivityRow } from 'components/channel/ActivityRow'
import { Table, TableHeaderRow, TableHeader } from 'components/styled/Table'

import { getChannelActivity } from 'state/channels'

interface Props {
  channel: ChannelState
}

export class ChannelActivityTable extends React.Component<Props, {}> {
  public constructor(props: any) {
    super(props)
  }

  public render() {
    return (
      <Table>
        <thead>
          <TableHeaderRow>
            <TableHeader align="left">Type</TableHeader>
            <TableHeader align="right">Your Delta</TableHeader>
            <TableHeader align="right">Your Balance</TableHeader>
            <TableHeader align="right">Their Delta</TableHeader>
            <TableHeader align="right">Their Balance</TableHeader>
          </TableHeaderRow>
        </thead>
        <tbody>
          {getChannelActivity(this.props.channel).map((activity, i) => (
            <ActivityRow
              op={activity.op}
              state={this.props.channel}
              key={i}
              pending={activity.pending}
              timestamp={activity.timestamp}
            />
          ))}
        </tbody>
      </Table>
    )
  }
}
