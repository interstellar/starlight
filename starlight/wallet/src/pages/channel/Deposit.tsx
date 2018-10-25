import * as React from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'

import { BtnSubmit } from 'pages/shared/Button'
import { RADICALRED } from 'pages/shared/Colors'
import { Heading } from 'pages/shared/Heading'
import { Icon } from 'pages/shared/Icon'
import { Hint, Input, Label, HelpBlock } from 'pages/shared/Input'
import { HorizontalLine } from 'pages/shared/HorizontalLine'
import { TransactionFee } from 'pages/shared/TransactionFee'
import { Total } from 'pages/shared/Total'
import { Unit, UnitContainer } from 'pages/shared/Unit'
import { ApplicationState } from 'types/schema'
import { deposit, getMyBalance } from 'state/channels'
import { ChannelState } from 'types/schema'
import { getWalletStroops } from 'state/wallet'
import { formatAmount, stroopsToLumens, lumensToStroops } from 'helpers/lumens'

const View = styled.div`
  padding: 25px;
`
const Form = styled.form`
  margin-top: 45px;
`
const ChannelName = styled.div`
  font-family: 'Nitti Grotesk';
  font-size: 24px;
  font-weight: 500;
  margin-bottom: 45px;
`

interface Props {
  channel: ChannelState
  availableBalance: number
  deposit: (id: string, amount: number) => void
  closeModal: () => void
}

interface State {
  amount: string
  channelName: string
  formErrors: {
    amount: boolean
  }
  loading: boolean
  showError: boolean
}

export class Deposit extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      amount: '',
      channelName: this.props.channel.CounterpartyAddress,
      formErrors: {
        amount: false,
      },
      loading: false,
      showError: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    const amount = this.amount()
    const fees = 100
    const total = amount && amount + fees

    return (
      <View>
        <Heading>Deposit to channel</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="Channel">Channel with</Label>
          <ChannelName>{this.state.channelName}</ChannelName>

          <div>
            <Label htmlFor="amount">Amount</Label>
            <Hint>
              <strong>
                {formatAmount(stroopsToLumens(this.props.availableBalance))}
              </strong>{' '}
              XLM available in account
            </Hint>
          </div>
          <UnitContainer>
            <Input
              value={this.state.amount}
              onBlur={() => {
                this.setState({
                  formErrors: {
                    amount:
                      !!this.state.amount &&
                      (parseFloat(this.state.amount) <= 0 ||
                        !this.walletHasSufficientBalance()),
                  },
                })
              }}
              onChange={e => {
                this.setState({ amount: e.target.value })
              }}
              type="number"
              name="amount"
              autoComplete="off"
              autoFocus
              error={this.state.formErrors.amount}
            />
            <Unit>XLM</Unit>
          </UnitContainer>

          <HelpBlock
            isShowing={
              amount !== undefined && !this.walletHasSufficientBalance()
            }
          >
            You only have{' '}
            {formatAmount(stroopsToLumens(this.props.availableBalance))} XLM
            available in your wallet.
          </HelpBlock>

          <HelpBlock isShowing={!!total && this.walletHasSufficientBalance()}>
            Once this is deposited, your new balance in this channel will be{' '}
            {formatAmount(
              stroopsToLumens(getMyBalance(this.props.channel) + (total || 0))
            )}{' '}
            XLM.
          </HelpBlock>

          <Label>Transaction Fee</Label>
          <TransactionFee>0.00001 XLM</TransactionFee>

          <HorizontalLine />

          <Label>Total Required</Label>
          {total && stroopsToLumens(total) !== 'NaN' ? (
            <Total>{formatAmount(stroopsToLumens(total))} XLM</Total>
          ) : (
            <Total>&mdash;</Total>
          )}

          {this.formatSubmitButton()}
        </Form>
      </View>
    )
  }

  private amount() {
    const lumensAmount = parseFloat(this.state.amount)
    if (isNaN(lumensAmount) || lumensAmount < 0) {
      return undefined
    }
    return lumensToStroops(lumensAmount)
  }

  private walletHasSufficientBalance() {
    const amount = this.amount()
    if (amount === undefined) {
      return false
    }
    return this.props.availableBalance >= amount
  }

  private formatSubmitButton() {
    if (this.state.loading) {
      return (
        <BtnSubmit disabled>
          Depositing <Icon className="fa-pulse" name="spinner" />
        </BtnSubmit>
      )
    } else if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Error depositing
        </BtnSubmit>
      )
    } else {
      return (
        <BtnSubmit disabled={!this.walletHasSufficientBalance()}>
          Deposit to channel
        </BtnSubmit>
      )
    }
  }

  private async handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    this.setState({ loading: true })

    const amount = this.amount()
    if (amount === undefined) {
      throw new Error('amount unexpectedly undefined')
    }
    if (!this.walletHasSufficientBalance()) {
      throw new Error('wallet unexpectedly does not have sufficient balance')
    }
    const ok = await this.props.deposit(this.props.channel.ID, amount)

    if (ok) {
      this.props.closeModal()
    } else {
      this.setState({ loading: false, showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    availableBalance: getWalletStroops(state),
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    deposit: (id: string, amount: number) => {
      return deposit(dispatch, id, amount)
    },
  }
}

export const ConnectedDeposit = connect<
  {},
  {},
  { channel: ChannelState; closeModal: () => void }
>(
  mapStateToProps,
  mapDispatchToProps
)(Deposit)
