import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'

import { ApplicationState } from 'schema'

import { ChannelRow } from 'components/channels/ChannelRow'
import { BarGraph } from 'components/graphs/BarGraph'
import { BtnHeading } from 'components/styled/Button'
import { CORNFLOWER, EBONYCLAY } from 'components/styled/Colors'
import { Container } from 'components/styled/Container'
import { AlertFlash } from 'components/styled/Flash'
import { Heading } from 'components/styled/Heading'
import { Modal } from 'components/styled/Modal'
import { ConnectedCreateChannel } from 'components/forms/CreateChannel'
import { Section, SectionHeading } from 'components/styled/Section'

import { ChannelState } from 'types'

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
    }

    this.openModal = this.openModal.bind(this)
    this.closeModal = this.closeModal.bind(this)
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

  public render() {
    return (
      <Container>
        { this.state.showFlash &&
          <AlertFlash>{this.props.location.state.message}</AlertFlash>
        }
        <Heading>Channels</Heading>
        <BtnHeading onClick={this.openModal}>Create channel</BtnHeading>
        <Modal isOpen={this.state.hasOpenModal} onClose={this.closeModal}>
          <ConnectedCreateChannel closeModal={() => this.closeModal()} />
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
          {this.props.channels.length > 0 ? this.props.channels.map(channel => (
            <ChannelRow channel={channel} key={channel.ID} />
          )) : (<Row>You haven't created any channels yet.</Row>)
          }
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
