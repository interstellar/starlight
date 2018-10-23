import * as React from 'react'
import styled from 'styled-components'
import { RouteComponentProps } from 'react-router-dom'
import { connect } from 'react-redux'
import { Redirect } from 'react-router'
import { Dispatch } from 'redux'

import { usernameToAddress } from 'helpers/account'

import { ApplicationState } from 'types/schema'
import { ChannelState } from 'types/schema'

import { fromNowPast } from 'helpers/moment'

import { ConnectedChannelActions } from 'pages/channel/ChannelActions'
import { ChannelActivityTable } from 'pages/channel/ChannelActivityTable'
import { BarGraph } from 'pages/shared/graphs/BarGraph'
import { CORNFLOWER, EBONYCLAY, RADICALRED } from 'pages/shared/Colors'
import { Container } from 'pages/shared/Container'
import { Detail, DetailLabel, DetailValue } from 'pages/shared/Detail'
import { Heading, HeadingContainer } from 'pages/shared/Heading'
import { Section, SectionHeading } from 'pages/shared/Section'
import { Status } from 'pages/shared/Status'

import {
  getMyBalance,
  getTheirBalance,
  getWithdrawalTime,
} from 'state/channels'
import { flash } from 'state/flash'

const ChannelHeading = styled(Heading)`
  color: ${CORNFLOWER};
`

const Subtitle = styled.span`
  color: ${EBONYCLAY}
  font-size: 16px;
`

interface Props extends RouteComponentProps<{ id: string }> {
  channel: ChannelState | undefined
  username: string
  setFlash: (message: string, color: string) => void
}

export class Channel extends React.Component<Props, {}> {
  public constructor(props: any) {
    super(props)
  }

  public componentDidMount() {
    if (this.props.channel) {
      document.title = `Channel with ${this.props.channel.CounterpartyAddress}`
    }
  }

  public render() {
    const channel = this.props.channel

    if (channel === undefined) {
      this.props.setFlash(
        `Channel not found: ${this.props.match.params.id}`,
        RADICALRED
      )
      return <Redirect to={{ pathname: '/channels' }} />
    }

    const isHost = channel.Role === 'Host'
    const sendCapacity = getMyBalance(channel)
    const receiveCapacity = getTheirBalance(channel)
    const status = channel.State.replace(/(.)([A-Z])/g, '$1 $2')
    const isOpen = channel.State !== 'Closed' && channel.State !== ''
    const showWithdrawalTime = [
      'AwaitingRatchet',
      'AwaitingSettlementMintime',
    ].includes(channel.State)

    return (
      <Container>
        <HeadingContainer>
          <ChannelHeading>
            Channel <Subtitle>with {channel.CounterpartyAddress}</Subtitle>
          </ChannelHeading>
          <ConnectedChannelActions channel={channel} />
        </HeadingContainer>

        <Section>
          <SectionHeading>Details</SectionHeading>
          <Detail>
            <DetailLabel>Status</DetailLabel>
            <DetailValue>
              <Status value={status}>{status}</Status>
            </DetailValue>
          </Detail>
          <Detail>
            <DetailLabel>Opened</DetailLabel>
            <DetailValue>{fromNowPast(channel.FundingTime)}</DetailValue>
          </Detail>
          <Detail>
            <DetailLabel>Opened by</DetailLabel>
            <DetailValue>
              {isHost
                ? `${usernameToAddress(this.props.username)} (you)`
                : channel.CounterpartyAddress}
            </DetailValue>
          </Detail>
          <Detail>
            <DetailLabel>Your Reserve</DetailLabel>
            <DetailValue>{isHost && isOpen ? '5.08 XLM' : '—'}</DetailValue>
          </Detail>
          {showWithdrawalTime && (
            <Detail>
              <DetailLabel>Withdrawal Time</DetailLabel>
              <DetailValue>{getWithdrawalTime(channel)}</DetailValue>
            </Detail>
          )}
        </Section>

        <Section>
          <SectionHeading>Capacity</SectionHeading>
          <BarGraph
            leftAmount={sendCapacity}
            rightAmount={receiveCapacity}
            leftColor={CORNFLOWER}
            rightColor={EBONYCLAY}
          />
        </Section>

        <Section>
          <SectionHeading>Activity</SectionHeading>
          <ChannelActivityTable channel={channel} />
        </Section>
      </Container>
    )
  }
}

const mapStateToProps = (
  state: ApplicationState,
  ownProps: RouteComponentProps<{ id: string }>
): { channel: ChannelState | undefined; username: string } => {
  return {
    channel: state.channels[ownProps.match.params.id],
    username: state.config.Username,
  }
}
const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    setFlash: (message: string, color: string) => {
      return flash.set(dispatch, message, color)
    },
  }
}
export const ConnectedChannel = connect<{}, {}, {}>(
  mapStateToProps,
  mapDispatchToProps
)(Channel)
