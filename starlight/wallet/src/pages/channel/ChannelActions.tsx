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

import { close, getMyBalance } from 'state/channels'

const TooltipBtn = styled(BtnHeading)`
  margin-left: 0;
`
const TooltipBtnWrapper = styled.span`
  margin-left: 10px;
`

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

              <TooltipBtnWrapper>
                <Tooltip
                  content="You have no money<br>
                    in this channel."
                  hover={getMyBalance(this.props.channel) <= 0}
                  direction="bottom"
                >
                  <TooltipBtn
                    disabled={getMyBalance(this.props.channel) <= 0}
                    onClick={() => this.openModal('send')}
                    color={SEAFOAM}
                  >
                    Send
                  </TooltipBtn>
                </Tooltip>
              </TooltipBtnWrapper>

              <TooltipBtnWrapper>
                <Tooltip
                  content="Only the party who opened<br>
                    the channel can deposit<br>
                    funds at this time."
                  hover={!isHost}
                  direction="bottom"
                >
                  <TooltipBtn
                    disabled={!isHost}
                    onClick={() => this.openModal('deposit')}
                  >
                    Deposit
                  </TooltipBtn>
                </Tooltip>
              </TooltipBtnWrapper>
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
