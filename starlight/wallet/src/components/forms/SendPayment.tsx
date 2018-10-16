import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'

import { BtnSubmit } from 'components/styled/Button'
import { Heading } from 'components/styled/Heading'
import { Icon } from 'components/styled/Icon'
import { Hint, Input, Label, HelpBlock } from 'components/styled/Input'
import { HorizontalLine } from 'components/styled/HorizontalLine'
import { Total } from 'components/styled/Total'
import { TransactionFee } from 'components/styled/TransactionFee'
import { Unit, UnitContainer } from 'components/styled/Unit'
import { RADICALRED, SEAFOAM } from 'components/styled/Colors'
import { ApplicationState, ChannelsState } from 'schema'
import { getWalletStroops, send } from 'state/wallet'
import {
  channelPay,
  getCounterpartyAccounts,
  getMyBalance,
  getTheirAccount,
} from 'state/channels'
import { stroopsToLumens, lumensToStroops } from 'lumens'
import { ChannelState } from 'schema'
const StrKey = require('stellar-base').StrKey

const View = styled.div`
  padding: 25px;
`
const Form = styled.form`
  margin-top: 45px;
`

interface State {
  Recipient: string
  Amount: string // TODO(croaky): number?
  showError: boolean
  loading: boolean
}

interface Props {
  AvailableBalance: number
  Channels: ChannelsState
  InitialRecipient?: string
  walletPay: (recipient: string, amount: number) => Promise<void>
  channelPay: (id: string, amount: number) => Promise<void>
  closeModal: () => void
  CounterpartyAccounts: { [id: string]: string }
}

export class SendPayment extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      Recipient: props.InitialRecipient || '',
      Amount: '',
      showError: false,
      loading: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    const validAmount = this.amount() !== undefined
    const hasChannel = this.destinationChannel() !== undefined
    const hasChannelWithInsufficientBalance =
      validAmount && hasChannel && !this.channelHasSufficientBalance()
    const hasChannelWithSufficientBalance =
      validAmount && hasChannel && !hasChannelWithInsufficientBalance
    const validRecipient = this.recipientIsValidAccount()
    const submittable =
      validAmount &&
      (hasChannelWithSufficientBalance ||
        (validRecipient && this.walletHasSufficientBalance()))
    const amount = this.amount()
    let total = amount !== undefined ? amount : undefined
    if (total !== undefined && !hasChannel) {
      total += 100
    }
    const focusOnRecipient = this.props.InitialRecipient === undefined
    const channelBalance = this.channelBalance()
    const displayChannelBtn =
      (hasChannel && !validAmount) || this.channelHasSufficientBalance()

    return (
      <View>
        <Heading>Send payment</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="Recipient">Recipient</Label>
          <Input
            value={this.state.Recipient}
            onChange={e => {
              this.setState({ Recipient: e.target.value })
            }}
            type="text"
            name="Recipient"
            autoComplete="off"
            autoFocus={focusOnRecipient}
          />
          {/* TODO: validate this is a Stellar account ID */}
          <div>
            <Label htmlFor="Amount">Amount</Label>
            <Hint>
              {hasChannel && (
                <span>
                  <strong>
                    {stroopsToLumens(
                      this.channelBalance() as number
                    ).toString()}{' '}
                    XLM
                  </strong>{' '}
                  available in channel;{' '}
                </span>
              )}
              <strong>
                {stroopsToLumens(this.props.AvailableBalance).toString()} XLM
              </strong>{' '}
              available in account
            </Hint>
          </div>
          <UnitContainer>
            <Input
              value={this.state.Amount}
              onChange={e => {
                this.setState({ Amount: e.target.value })
              }}
              type="number"
              name="Amount"
              autoComplete="off"
              autoFocus={!focusOnRecipient}
            />
            {/* TODO: validate this is a number, and not more than the wallet balance */}
            <Unit>XLM</Unit>
          </UnitContainer>
          <HelpBlock isShowing={!hasChannel && validRecipient}>
            You do not have a channel open with this recipient. Open a channel
            or proceed to send this payment from your account on the Stellar
            network.
          </HelpBlock>
          <HelpBlock
            isShowing={
              hasChannelWithInsufficientBalance &&
              this.walletHasSufficientBalance()
            }
          >
            You only have {stroopsToLumens(this.channelBalance() || 0)} XLM
            available in this channel. The entire payment will occur on the
            Stellar network from your account instead.
          </HelpBlock>
          <HelpBlock
            isShowing={
              validRecipient &&
              validAmount &&
              !hasChannelWithSufficientBalance &&
              !this.walletHasSufficientBalance()
            }
          >
            {channelBalance === undefined ||
            this.props.AvailableBalance > channelBalance
              ? `You only have ${stroopsToLumens(
                  this.props.AvailableBalance
                )} XLM available in your wallet.`
              : `You only have ${stroopsToLumens(
                  channelBalance
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
          {total !== undefined ? (
            <Total>{stroopsToLumens(total)} XLM</Total>
          ) : (
            <Total>&mdash;</Total>
          )}

          {this.formatSubmitButton(displayChannelBtn, submittable)}
        </Form>
      </View>
    )
  }

  private recipientIsValidAccount() {
    const validAccountId = StrKey.isValidEd25519PublicKey(this.state.Recipient)

    return (
      this.destinationChannel() || validAccountId // TODO: || is valid Stellar address
    )
  }

  private destinationChannel(): ChannelState | undefined {
    const recipient = this.state.Recipient
    const channels = this.props.Channels

    if (channels[recipient] && channels[recipient].State !== 'Closed') {
      return channels[recipient]
    } else {
      return this.lookupChannelByAccountId(recipient, channels)
    }
  }

  private lookupChannelByAccountId(accountId: string, channels: ChannelsState) {
    const counterpartyAccounts = this.props.CounterpartyAccounts
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
    const lumensAmount = parseFloat(this.state.Amount)
    if (isNaN(lumensAmount)) {
      return undefined
    }
    // TODO: fix floating point problems
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
      throw new Error('invalid amount: ' + this.state.Amount)
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
        ok = await this.props.walletPay(this.state.Recipient, amount)
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
    return this.props.AvailableBalance >= amount
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    AvailableBalance: getWalletStroops(state),
    Channels: state.channels,
    CounterpartyAccounts: getCounterpartyAccounts(state),
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
  { closeModal: () => void; InitialRecipient?: string }
>(
  mapStateToProps,
  mapDispatchToProps
)(SendPayment)
