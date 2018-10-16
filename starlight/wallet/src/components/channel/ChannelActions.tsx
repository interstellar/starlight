import * as React from 'react'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'

import { ChannelState } from 'schema'

import { ConnectedCreateChannel } from 'components/forms/CreateChannel'
import { ConnectedDeposit } from 'components/forms/Deposit'
import { ConnectedSendPayment } from 'components/forms/SendPayment'

import { BtnHeading } from 'components/styled/Button'
import { RADICALRED, SEAFOAM } from 'components/styled/Colors'
import { DisabledBtnHover } from 'components/styled/DisabledBtnHover'
import { ActionContainer } from 'components/styled/Heading'
import { Modal } from 'components/styled/Modal'

import { close, getMyBalance } from 'state/channels'

interface Props {
  channel: ChannelState
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
    const isHost = this.props.channel.Role === 'Host'

    return (
      <div>
        <ActionContainer>
          {this.props.channel.State === 'Closed' ? (
            <span>
              <BtnHeading onClick={() => this.openModal('open')}>
                Open
              </BtnHeading>
            </span>
          ) : (
            <span>
              <BtnHeading
                onClick={() => this.props.closeChannel(this.props.channel.ID)}
                color={RADICALRED}
              >
                Close
              </BtnHeading>

              <BtnHeading
                disabled={getMyBalance(this.props.channel) <= 0}
                onClick={() => this.openModal('send')}
                color={SEAFOAM}
              >
                Send
              </BtnHeading>

              <DisabledBtnHover
                content="Only the person who opened
                  <br> the channel can deposit
                  <br> funds at this time."
                disable={isHost}
              >
                <BtnHeading
                  disabled={!isHost}
                  onClick={() => this.openModal('deposit')}
                >
                  Deposit
                </BtnHeading>
              </DisabledBtnHover>
            </span>
          )}
        </ActionContainer>

        <Modal isOpen={this.hasOpenModal('deposit')} onClose={this.closeModal}>
          <ConnectedDeposit
            channel={this.props.channel}
            closeModal={this.closeModal}
          />
        </Modal>
        <Modal isOpen={this.hasOpenModal('send')} onClose={this.closeModal}>
          <ConnectedSendPayment
            InitialRecipient={this.props.channel.CounterpartyAddress}
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
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
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
