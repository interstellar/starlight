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
import { ApplicationState } from 'schema'
import { deposit } from 'state/channels'
import { ChannelState } from 'schema'
import { getWalletStroops } from 'state/wallet'
import { stroopsToLumens, lumensToStroops } from 'helpers/lumens'

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
  AvailableBalance: number
  deposit: (id: string, amount: number) => void
  closeModal: () => void
}

interface State {
  Amount: string
  ChannelName: string
  showError: boolean
  loading: boolean
}

export class Deposit extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      Amount: '',
      ChannelName: this.props.channel.CounterpartyAddress,
      showError: false,
      loading: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    const amount = this.amount()

    return (
      <View>
        <Heading>Deposit to channel</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="Channel">Channel</Label>
          <ChannelName>{this.state.ChannelName}</ChannelName>

          <div>
            <Label htmlFor="Amount">Amount</Label>
            <Hint>
              <strong>{stroopsToLumens(this.props.AvailableBalance)}</strong>{' '}
              XLM available in account
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
              autoFocus
            />
            <Unit>XLM</Unit>
          </UnitContainer>

          <HelpBlock
            isShowing={
              amount !== undefined && !this.walletHasSufficientBalance()
            }
          >
            You only have {stroopsToLumens(this.props.AvailableBalance)} XLM
            available in your wallet.
          </HelpBlock>

          <Label>Transaction Fee</Label>
          <TransactionFee>0.00001 XLM</TransactionFee>

          <HorizontalLine />

          <Label>Total Required</Label>
          {amount !== undefined ? (
            <Total>{stroopsToLumens(amount + 100)} XLM</Total>
          ) : (
            <Total>&mdash;</Total>
          )}

          {this.formatSubmitButton()}
        </Form>
      </View>
    )
  }

  private amount() {
    const lumensAmount = parseFloat(this.state.Amount)
    if (isNaN(lumensAmount)) {
      return undefined
    }
    return lumensToStroops(lumensAmount)
  }

  private walletHasSufficientBalance() {
    const amount = this.amount()
    if (amount === undefined) {
      return false
    }
    return this.props.AvailableBalance >= amount
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
    AvailableBalance: getWalletStroops(state),
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
