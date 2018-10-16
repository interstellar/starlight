import * as React from 'react'
import styled from 'styled-components'
import { Link } from 'react-router-dom'

import { ChannelState } from 'schema'

import { MiniBarGraph } from 'components/graphs/MiniBarGraph'
import { DUSTYGRAY, EBONYCLAY, WILDSAND_LIGHT } from 'components/styled/Colors'
import { Status } from 'components/styled/Status'

import { getMyBalance, getTheirBalance } from 'state/channels'

const Address = styled.div`
  color: ${EBONYCLAY};
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 500;
`
const AddressCell = styled.div`
  flex: 2;
  padding-left: 40px;
`
const Arrow = styled.span`
  color: ${DUSTYGRAY};
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 700;
  text-align: right;
`
const ArrowCell = styled.div`
  flex: 1;
  padding-right: 40px;
  text-align: right;
`
const BarGraphCell = styled.div`
  flex: 4;
`
const Row = styled.div`
  align-items: center;
  display: flex;

  &:hover {
    background-color: ${WILDSAND_LIGHT};
  }
`
const RowLink = styled(Link)`
  text-decoration: none;
`
const StatusCell = styled.div`
  flex: 1;
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  font-weight: 500;
  padding-left: 40px;
`

interface Props {
  channel: ChannelState
}

export class ChannelRow extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    const channel = this.props.channel
    const ID = channel.CounterpartyAddress
    const status = channel.State.replace(/(.)([A-Z])/g, '$1 $2')

    return (
      <RowLink to={'/channel/' + ID}>
        <Row>
          <AddressCell>
            <Address>{this.props.channel.CounterpartyAddress}</Address>
          </AddressCell>
          <BarGraphCell>
            <MiniBarGraph
              leftAmount={getMyBalance(channel)}
              rightAmount={getTheirBalance(channel)}
            />
          </BarGraphCell>
          <StatusCell>
            <Status value={status}>{status}</Status>
          </StatusCell>
          <ArrowCell>
            <Arrow>&rarr;</Arrow>
          </ArrowCell>
        </Row>
      </RowLink>
    )
  }
}
