import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'

import { ChannelState } from 'types/schema'

import { ConnectedCreateChannel } from 'pages/shared/forms/CreateChannel'
import { ConnectedDeposit } from 'pages/channel/Deposit'
import { ConnectedSendPayment } from 'pages/shared/forms/SendPayment'

import { BtnHeading } from 'pages/shared/Button'
import { RADICALRED, SEAFOAM } from 'pages/shared/Colors'
import { ActionContainer } from 'pages/shared/Heading'
import { Modal } from 'pages/shared/Modal'
import { Tooltip } from 'pages/shared/Tooltip'

import { cancel, close, getMyBalance } from 'state/channels'

const TooltipBtn = styled(BtnHeading)`
  margin-left: 0;
`
const TooltipBtnWrapper = styled.span`
  margin-left: 10px;
`

interface Props {
  channel: ChannelState
  cancelOpenChannel: (id: string) => void
  closeChannel: (id: string) => undefined
}

interface State {
  openedModalName: string
}

export class ChannelActions extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      openedModalName: '',
    }

    this.openModal = this.openModal.bind(this)
    this.closeModal = this.closeModal.bind(this)
  }

  private openModal(name: string) {
    this.setState({ openedModalName: name })
  }

  private hasOpenModal(name: string) {
    return this.state.openedModalName === name
  }

  private closeModal() {
    this.setState({ openedModalName: '' })
  }

  public render() {
    return (
      <div>
        <ActionContainer>{this.buttonsForChannelState()}</ActionContainer>

        <Modal isOpen={this.hasOpenModal('deposit')} onClose={this.closeModal}>
          <ConnectedDeposit
            channel={this.props.channel}
            closeModal={this.closeModal}
          />
        </Modal>
        <Modal isOpen={this.hasOpenModal('send')} onClose={this.closeModal}>
          <ConnectedSendPayment
            initialRecipient={this.props.channel.CounterpartyAddress}
            closeModal={this.closeModal}
          />
        </Modal>
        <Modal isOpen={this.hasOpenModal('open')} onClose={this.closeModal}>
          <ConnectedCreateChannel
            closeModal={() => this.closeModal()}
            prefill={{ counterparty: this.props.channel.CounterpartyAddress }}
          />
        </Modal>
      </div>
    )
  }

  private buttonsForChannelState() {
    const channelState = this.props.channel.State

    if (channelState === 'Closed') {
      return <span>{this.openChannelBtn()}</span>
    } else if (channelState === 'ChannelProposed') {
      return <span>{this.cancelOpenChannelBtn()}</span>
    } else {
      return (
        <span>
          {this.closeChannelBtn()}
          {this.sendBtn()}

          {this.depositBtn()}
        </span>
      )
    }
  }

  private openChannelBtn() {
    return <BtnHeading onClick={() => this.openModal('open')}>Open</BtnHeading>
  }

  private cancelOpenChannelBtn() {
    return (
      <BtnHeading
        onClick={() => this.props.cancelOpenChannel(this.props.channel.ID)}
      >
        Cancel
      </BtnHeading>
    )
  }

  private closeChannelBtn() {
    return (
      <BtnHeading
        disabled={this.props.channel.State !== 'Open'}
        onClick={() => this.props.closeChannel(this.props.channel.ID)}
        color={RADICALRED}
      >
        Close
      </BtnHeading>
    )
  }

  private sendBtn() {
    return (
      <TooltipBtnWrapper>
        <Tooltip
          content="You have no money<br>
          in this channel."
          hover={getMyBalance(this.props.channel) <= 0}
          direction="bottom"
        >
          <TooltipBtn
            disabled={
              this.props.channel.State !== 'Open' ||
              getMyBalance(this.props.channel) <= 0
            }
            onClick={() => this.openModal('send')}
            color={SEAFOAM}
          >
            Send
          </TooltipBtn>
        </Tooltip>
      </TooltipBtnWrapper>
    )
  }

  private depositBtn() {
    const isHost = this.props.channel.Role === 'Host'

    return (
      <TooltipBtnWrapper>
        <Tooltip
          content="Only the party who opened<br>
          the channel can deposit funds."
          hover={!isHost}
          direction="bottom"
        >
          <TooltipBtn
            disabled={this.props.channel.State !== 'Open' || !isHost}
            onClick={() => this.openModal('deposit')}
          >
            Deposit
          </TooltipBtn>
        </Tooltip>
      </TooltipBtnWrapper>
    )
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    cancelOpenChannel: (id: string) => {
      cancel(dispatch, id)
    },
    closeChannel: (id: string) => {
      close(dispatch, id)
    },
  }
}

export const ConnectedChannelActions = connect<
  {},
  {},
  { channel: ChannelState }
>(
  null,
  mapDispatchToProps
)(ChannelActions)
