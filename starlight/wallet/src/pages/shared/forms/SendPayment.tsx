import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'

import { validRecipientAccount } from 'helpers/account'
import { formatAmount, stroopsToLumens, lumensToStroops } from 'helpers/lumens'

import { BtnSubmit } from 'pages/shared/Button'
import { Heading } from 'pages/shared/Heading'
import { Icon } from 'pages/shared/Icon'
import { Hint, Input, Label, HelpBlock } from 'pages/shared/Input'
import { HorizontalLine } from 'pages/shared/HorizontalLine'
import { Total } from 'pages/shared/Total'
import { TransactionFee } from 'pages/shared/TransactionFee'
import { Unit, UnitContainer } from 'pages/shared/Unit'
import { RADICALRED, SEAFOAM } from 'pages/shared/Colors'
import { ApplicationState, ChannelsState } from 'types/schema'
import { getWalletStroops, send } from 'state/wallet'

import {
  channelPay,
  getCounterpartyAccounts,
  getMyBalance,
  getTheirAccount,
} from 'state/channels'
import { ChannelState } from 'types/schema'

const View = styled.div`
  padding: 25px;
`
const Form = styled.form`
  margin-top: 45px;
`

interface State {
  amount: string
  formErrors: {
    amount: boolean,
    recipient: boolean
  },
  loading: boolean
  recipient: string
  showError: boolean
}

interface Props {
  availableBalance: number
  channels: ChannelsState
  initialRecipient?: string
  walletPay: (recipient: string, amount: number) => Promise<void>
  channelPay: (id: string, amount: number) => Promise<void>
  closeModal: () => void
  counterpartyAccounts: { [id: string]: string }
  username: string
}

export class SendPayment extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      amount: '',
      formErrors: {
        amount: false,
        recipient: false,
      },
      loading: false,
      recipient: props.initialRecipient || '',
      showError: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    const validAmount = this.amount() !== undefined
    const hasChannel = this.destinationChannel() !== undefined
    const hasChannelWithSufficientBalance =
      validAmount && hasChannel && this.channelHasSufficientBalance()
    const submittable =
      validAmount &&
      (hasChannelWithSufficientBalance ||
        (this.recipientIsValid() && this.walletHasSufficientBalance()))
    const amount = this.amount()
    let total = amount
    if (total !== undefined && !hasChannelWithSufficientBalance) {
      total += 100
    }
    const focusOnRecipient = this.props.initialRecipient === undefined
    const channelBalance = this.channelBalance()
    const displayChannelBtn =
      (hasChannel && !validAmount) || this.channelHasSufficientBalance()

    return (
      <View>
        <Heading>Send payment</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="recipient">Recipient</Label>
          <Input
            value={this.state.recipient}
            onBlur={() => {
              this.setState({
                formErrors: {
                  amount: this.state.formErrors.amount,
                  recipient: !!this.state.recipient && !this.recipientIsValid(),
                },
              })
            }}
            onChange={e => {
              this.setState({ recipient: e.target.value })
            }}
            type="text"
            name="recipient"
            autoComplete="off"
            autoFocus={focusOnRecipient}
            error={this.state.formErrors.recipient}
          />
          {/* TODO: validate this is a Stellar account ID */}
          <div>
            <Label htmlFor="amount">Amount</Label>
            <Hint>
              {hasChannel && (
                <span>
                  <strong>
                    {formatAmount(
                      stroopsToLumens(this.channelBalance() as number)
                    )}{' '}
                    XLM
                  </strong>{' '}
                  available in channel;{' '}
                </span>
              )}
              <strong>
                {formatAmount(
                  stroopsToLumens(this.props.availableBalance).toString()
                )}{' '}
                XLM
              </strong>{' '}
              available in account
            </Hint>
          </div>
          <UnitContainer>
            <Input
              value={this.state.amount}
              onBlur={() => {
                this.setState({
                  formErrors: {
                    amount: !!this.state.amount && (
                      parseFloat(this.state.amount) <= 0 ||
                      !validAmount ||
                      !this.walletHasSufficientBalance()
                    ),
                    recipient: this.state.formErrors.recipient,
                  },
                })
              }}
              onChange={e => {
                this.setState({ amount: e.target.value })
              }}
              type="number"
              name="amount"
              autoComplete="off"
              autoFocus={!focusOnRecipient}
              error={this.state.formErrors.amount}
            />
            {/* TODO: validate this is a number, and not more than the wallet balance */}
            <Unit>XLM</Unit>
          </UnitContainer>
          <HelpBlock isShowing={!hasChannel && this.recipientIsValid()}>
            You do not have a channel open with this recipient. Open a channel
            or proceed to send this payment from your account on the Stellar
            network.
          </HelpBlock>
          <HelpBlock
            isShowing={
              !hasChannelWithSufficientBalance &&
              this.walletHasSufficientBalance()
            }
          >
            You only have{' '}
            {formatAmount(stroopsToLumens(this.channelBalance() || 0))} XLM
            available in this channel. The entire payment will occur on the
            Stellar network from your account instead.
          </HelpBlock>
          <HelpBlock
            isShowing={
              this.recipientIsValid() &&
              validAmount &&
              !hasChannelWithSufficientBalance &&
              !this.walletHasSufficientBalance()
            }
          >
            {channelBalance === undefined ||
            this.props.availableBalance > channelBalance
              ? `You only have ${formatAmount(
                  stroopsToLumens(this.props.availableBalance)
                )} XLM available in your wallet.`
              : `You only have ${formatAmount(
                  stroopsToLumens(channelBalance)
                )} XLM available in this channel.`}
          </HelpBlock>
          <Label>Transaction Fee</Label>
          {(!validAmount && hasChannel) ||
          this.channelHasSufficientBalance() ? (
            <TransactionFee>&mdash;</TransactionFee>
          ) : (
            <TransactionFee>0.00001 XLM</TransactionFee>
          )}
          <HorizontalLine />
          <Label>Total</Label>
          {total && stroopsToLumens(total) !== 'NaN' ? (
            <Total>{stroopsToLumens(total)} XLM</Total>
          ) : (
            <Total>&mdash;</Total>
          )}

          {this.formatSubmitButton(displayChannelBtn, submittable)}
        </Form>
      </View>
    )
  }

  private recipientIsValid() {
    return (
      this.destinationChannel() ||
      validRecipientAccount(this.props.username, this.state.recipient)
    )
  }

  private destinationChannel(): ChannelState | undefined {
    const recipient = this.state.recipient
    const channels = this.props.channels

    if (channels[recipient] && channels[recipient].State !== 'Closed') {
      return channels[recipient]
    } else {
      return this.lookupChannelByAccountId(recipient, channels)
    }
  }

  private lookupChannelByAccountId(accountId: string, channels: ChannelsState) {
    const counterpartyAccounts = this.props.counterpartyAccounts
    const chanId = counterpartyAccounts[accountId]

    if (chanId !== undefined) {
      const channel = channels[chanId]
      if (channel.State !== 'Closed') {
        return channel
      }
    }
  }

  private channelBalance() {
    const destinationChannel = this.destinationChannel()
    return destinationChannel && getMyBalance(destinationChannel)
  }

  private amount() {
    const lumensAmount = parseFloat(this.state.amount)
    if (isNaN(lumensAmount) || lumensAmount < 0) {
      return undefined
    }
    return lumensToStroops(lumensAmount)
  }

  private formatSubmitButton(displayChannelBtn: boolean, submittable: boolean) {
    if (this.state.loading) {
      return (
        <BtnSubmit disabled>
          Sending <Icon className="fa-pulse" name="spinner" />
        </BtnSubmit>
      )
    } else if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Payment not sent
        </BtnSubmit>
      )
    } else {
      if (displayChannelBtn) {
        return <BtnSubmit disabled={!submittable}>Send via channel</BtnSubmit>
      } else {
        return (
          <BtnSubmit disabled={!submittable} color={SEAFOAM}>
            Send on Stellar
          </BtnSubmit>
        )
      }
    }
  }

  private async handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    this.setState({ loading: true })

    const amount = this.amount()
    if (amount === undefined) {
      throw new Error('invalid amount: ' + this.state.amount)
    }

    let ok
    const channel = this.destinationChannel()
    if (channel && this.channelHasSufficientBalance()) {
      ok = await this.props.channelPay(channel.ID, amount)
    } else {
      if (!this.walletHasSufficientBalance()) {
        throw new Error('wallet has insufficient balance')
      }
      if (channel) {
        ok = await this.props.walletPay(getTheirAccount(channel), amount)
      } else {
        ok = await this.props.walletPay(this.state.recipient, amount)
      }
    }

    if (ok) {
      this.props.closeModal()
      // TODO: drop you into channel view when successful
    } else {
      this.setState({ loading: false, showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }

  private channelHasSufficientBalance() {
    const balance = this.channelBalance()
    if (balance === undefined) {
      return false
    }
    const amount = this.amount()
    if (amount === undefined) {
      return false
    }
    return balance >= amount
  }

  private walletHasSufficientBalance() {
    const amount = this.amount()
    if (amount === undefined) {
      return false
    }
    return this.props.availableBalance >= amount
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    availableBalance: getWalletStroops(state),
    channels: state.channels,
    counterpartyAccounts: getCounterpartyAccounts(state),
    username: state.config.Username,
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    walletPay: (recipient: string, amount: number) => {
      return send(dispatch, recipient, amount)
    },
    channelPay: (id: string, amount: number) => {
      return channelPay(dispatch, id, amount)
    },
  }
}

export const ConnectedSendPayment = connect<
  {},
  {},
  { closeModal: () => void; initialRecipient?: string }
>(
  mapStateToProps,
  mapDispatchToProps
)(SendPayment)
