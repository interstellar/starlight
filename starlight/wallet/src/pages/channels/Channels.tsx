import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { Redirect } from 'react-router'

import { ApplicationState } from 'types/schema'

import { ChannelRow } from 'pages/channels/ChannelRow'
import { BarGraph } from 'pages/shared/graphs/BarGraph'
import { BtnHeading } from 'pages/shared/Button'
import { CORNFLOWER, EBONYCLAY } from 'pages/shared/Colors'
import { Container } from 'pages/shared/Container'
import { AlertFlash } from 'pages/shared/Flash'
import { Heading } from 'pages/shared/Heading'
import { Modal } from 'pages/shared/Modal'
import { ConnectedCreateChannel } from 'pages/shared/forms/CreateChannel'
import { Section, SectionHeading } from 'pages/shared/Section'

import { ChannelState } from 'types/schema'

import {
  getChannels,
  getTotalChannelBalance,
  getTotalChannelCounterpartyBalance,
} from 'state/channels'

const ChannelListSection = styled(Section)`
  padding: 30px 0;
`
const Row = styled.div`
  padding-left: 40px;
`

interface Props {
  channels: ChannelState[]
  location: any
  totalChannelBalance: number
  totalChannelCounterpartyBalance: number
  username: string
}

interface State {
  hasOpenModal: boolean
  showFlash: boolean
  timer?: number
  redirectTo: string
}

export class Channels extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      hasOpenModal: false,
      showFlash: !!this.props.location.state,
      timer: window.setTimeout(() => {
        this.setState({ showFlash: false })
      }, 3000),
      redirectTo: '',
    }

    this.openModal = this.openModal.bind(this)
    this.closeModal = this.closeModal.bind(this)
    this.redirect = this.redirect.bind(this)
  }

  public componentDidMount() {
    document.title = `Channels - ${this.props.username}`
  }

  public componentWillUnmount() {
    clearTimeout(this.state.timer)
  }

  private openModal() {
    this.setState({ hasOpenModal: true })
  }

  private closeModal() {
    this.setState({ hasOpenModal: false })
  }

  private redirect(account: string) {
    this.setState({ redirectTo: account })
  }

  public render() {
    if (this.state.redirectTo) {
      return <Redirect to={`/channel/${this.state.redirectTo}`} />
    }

    return (
      <Container>
        {this.state.showFlash && (
          <AlertFlash>{this.props.location.state.message}</AlertFlash>
        )}
        <Heading>Channels</Heading>
        <BtnHeading onClick={this.openModal}>Create channel</BtnHeading>
        <Modal isOpen={this.state.hasOpenModal} onClose={this.closeModal}>
          <ConnectedCreateChannel
            closeModal={() => this.closeModal()}
            redirect={this.redirect}
          />
        </Modal>
        <Section>
          <SectionHeading>Capacity</SectionHeading>
          <BarGraph
            leftAmount={this.props.totalChannelBalance}
            rightAmount={this.props.totalChannelCounterpartyBalance}
            leftColor={CORNFLOWER}
            rightColor={EBONYCLAY}
          />
        </Section>
        <ChannelListSection>
          {this.props.channels.length > 0 ? (
            this.props.channels.map(channel => (
              <ChannelRow channel={channel} key={channel.ID} />
            ))
          ) : (
            <Row>You haven't created any channels yet.</Row>
          )}
        </ChannelListSection>
      </Container>
    )
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    channels: getChannels(state),
    totalChannelBalance: getTotalChannelBalance(state),
    totalChannelCounterpartyBalance: getTotalChannelCounterpartyBalance(state),
    username: state.config.Username,
  }
}

export const ConnectedChannels = connect(
  mapStateToProps,
  {}
)(Channels)
