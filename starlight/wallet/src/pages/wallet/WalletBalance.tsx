import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'

import { ApplicationState } from 'types/schema'

import { BarGraph } from 'pages/shared/graphs/BarGraph'
import {
  ALTO,
  CORNFLOWER,
  RIVERBED,
  DUSTYGRAY,
  SEAFOAM_LIGHT,
} from 'pages/shared/Colors'

import { getWalletStroops } from 'state/wallet'
import { formatAmount, lumensToStroops, stroopsToLumens } from 'helpers/lumens'

import {
  getNumberOfOpenHostChannels,
  getTotalChannelBalance,
} from 'state/channels'

const AvailableWrapper = styled.div`
  display: flex;
`
const BalanceContainer = styled.div`
  border-right: ${ALTO} 1px solid;
  padding-right: 50px;
  width: auto;
`
const Balance = styled.div`
  color: ${RIVERBED};
  font-family: 'Nitti Grotesk';
  font-size: 36px;
  font-weight: 700;
`
const GraphContainer = styled.div`
  flex-grow: 2;
  padding-left: 50px;
`
const Reserve = styled.div`
  color: ${DUSTYGRAY};
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  font-weight: 700;
`

interface Props {
  channelBalance: number
  reserve: number
  walletBalance: number
}

export class WalletBalance extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  private total() {
    return (
      this.props.walletBalance + this.props.channelBalance + this.props.reserve
    )
  }

  public render() {
    return (
      <AvailableWrapper>
        <BalanceContainer>
          <Balance>
            {formatAmount(stroopsToLumens(this.total(), { short: true }))} XLM
          </Balance>
          {!!this.props.reserve && (
            <Reserve>
              {formatAmount(
                stroopsToLumens(
                  this.props.walletBalance + this.props.channelBalance,
                  { short: true }
                )
              )} XLM Available +{' '}
              {formatAmount(
                stroopsToLumens(this.props.reserve, { short: true })
              )} XLM Reserve
            </Reserve>
          )}
        </BalanceContainer>
        <GraphContainer>
          <BarGraph
            leftLabel="Account"
            leftAmount={this.props.walletBalance}
            rightAmount={this.props.channelBalance}
            rightLabel="Channels"
            leftColor={SEAFOAM_LIGHT}
            rightColor={CORNFLOWER}
          />
        </GraphContainer>
      </AvailableWrapper>
    )
  }
}

const mapStateToProps = (state: ApplicationState) => {
  const walletAccountExists = getWalletStroops(state) !== 0
  const walletReserve = walletAccountExists ? lumensToStroops(1) : 0

  return {
    reserve:
      getNumberOfOpenHostChannels(state) * lumensToStroops(5) + walletReserve,
    walletBalance: getWalletStroops(state) - walletReserve,
    channelBalance: getTotalChannelBalance(state),
  }
}
export const ConnectedWalletBalance = connect(
  mapStateToProps,
  {}
)(WalletBalance)
