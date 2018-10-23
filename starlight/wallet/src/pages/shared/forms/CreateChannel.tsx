import * as React from 'react'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import styled from 'styled-components'

import { validRecipientAccount } from 'helpers/account'
import { formatAmount, stroopsToLumens, lumensToStroops } from 'helpers/lumens'

import { BtnSubmit } from 'pages/shared/Button'
import { CORNFLOWER, RADICALRED } from 'pages/shared/Colors'
import { Heading } from 'pages/shared/Heading'
import { Icon } from 'pages/shared/Icon'
import { Hint, Input, Label, HelpBlock } from 'pages/shared/Input'
import { HorizontalLine } from 'pages/shared/HorizontalLine'
import { Tooltip } from 'pages/shared/Tooltip'
import { Total } from 'pages/shared/Total'
import { Unit, UnitContainer } from 'pages/shared/Unit'

import { ApplicationState } from 'types/schema'
import { getWalletStroops } from 'state/wallet'
import { createChannel } from 'state/channels'

const Amount = styled.div`
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 45px;
  text-transform: uppercase;
`
const Form = styled.form`
  margin-top: 45px;
`
const HalfWidth = styled.div`
  display: inline-block;
  width: 50%;
`
const InfoIcon = styled(Icon)`
  color: ${CORNFLOWER};
  cursor: pointer;

  &:hover {
    opacity: 0.8;
  }
`
const View = styled.div`
  padding: 25px;
`

interface Props {
  availableBalance: number
  closeModal: () => void
  createChannel: (recipient: string, initialDeposit: number) => void
  prefill?: { counterparty: string }
  redirect?: (account: string) => void
  username: string
}

interface State {
  counterparty: string
  initialDeposit: string
  showError: boolean
  showTooltip: boolean
  loading: boolean
  formErrors: {
    deposit: boolean
    counterparty: boolean
  }
}

export class CreateChannel extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      counterparty: this.props.prefill ? this.props.prefill.counterparty : '',
      initialDeposit: '',
      showError: false,
      showTooltip: false,
      loading: false,
      formErrors: {
        deposit: false,
        counterparty: false,
      },
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    const initialDeposit = this.amount()
    const reserve = 50800000
    const fee = 700
    const total = initialDeposit && initialDeposit + reserve + fee

    return (
      <View>
        <Heading>Create channel</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="counterparty">counterparty</Label>
          <Input
            value={this.state.counterparty}
            onBlur={() => {
              if (
                this.state.counterparty &&
                !validRecipientAccount(
                  this.props.username,
                  this.state.counterparty
                )
              ) {
                this.setState({
                  formErrors: {
                    deposit: this.state.formErrors.deposit,
                    counterparty: true,
                  },
                })
              } else {
                this.setState({
                  formErrors: {
                    deposit: this.state.formErrors.deposit,
                    counterparty: false,
                  },
                })
              }
            }}
            onChange={e => {
              this.setState({ counterparty: e.target.value })
            }}
            type="text"
            name="counterparty"
            autoComplete="off"
            autoFocus={!this.state.counterparty}
            error={this.state.formErrors.counterparty}
          />

          <Label htmlFor="initialDeposit">Initial Deposit</Label>
          <Hint>
            <strong>
              {formatAmount(stroopsToLumens(this.props.availableBalance))} XLM
            </strong>{' '}
            available in account
          </Hint>
          <UnitContainer>
            <Input
              value={this.state.initialDeposit}
              onBlur={() => {
                if (
                  this.state.initialDeposit &&
                  !this.walletHasSufficientBalance()
                ) {
                  this.setState({
                    formErrors: {
                      deposit: true,
                      counterparty: this.state.formErrors.counterparty,
                    },
                  })
                } else {
                  this.setState({
                    formErrors: {
                      deposit: false,
                      counterparty: this.state.formErrors.counterparty,
                    },
                  })
                }
              }}
              onChange={e => {
                this.setState({ initialDeposit: e.target.value })
              }}
              type="number"
              name="initialDeposit"
              autoComplete="off"
              autoFocus={!!this.state.counterparty}
              error={this.state.formErrors.deposit}
            />
            <Unit>XLM</Unit>
          </UnitContainer>

          <HelpBlock
            isShowing={!!this.amount() && !this.walletHasSufficientBalance()}
          >
            You only have{' '}
            {formatAmount(stroopsToLumens(this.props.availableBalance))} XLM
            available in your wallet.
          </HelpBlock>

          <HalfWidth>
            <Label>Transaction Fee</Label>
            <Amount>0.00007 XLM</Amount>
          </HalfWidth>

          <HalfWidth>
            <Label>
              Channel Reserve{' '}
              <Tooltip
                content="This a required minimum balance for a<br>
                Starlight payment channel. It cannot be<br>
                spent while the channel is open, but will<br>
                be returned when the channel is closed,<br>
                after subtracting fees for the closing<br>
                transactions."
                hover
              >
                <InfoIcon name="info-circle" />
              </Tooltip>
            </Label>
            <Amount>{stroopsToLumens(reserve)} XLM</Amount>
          </HalfWidth>

          <HorizontalLine />

          <Label>Total Required</Label>
          <Total>
            {!total || !this.walletHasSufficientBalance()
              ? '—'
              : stroopsToLumens(total) + ' XLM'}
          </Total>
          {this.formatSubmitButton()}
        </Form>
      </View>
    )
  }

  private amount() {
    const amountFloat = parseFloat(this.state.initialDeposit)
    if (isNaN(amountFloat) || amountFloat < 0) {
      return undefined
    }
    return lumensToStroops(amountFloat)
  }

  private formatSubmitButton() {
    if (this.state.loading) {
      return (
        <BtnSubmit disabled>
          Opening <Icon className="fa-pulse" name="spinner" />
        </BtnSubmit>
      )
    } else if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Error opening channel
        </BtnSubmit>
      )
    } else {
      return <BtnSubmit disabled={!this.formIsValid()}>Open channel</BtnSubmit>
    }
  }

  private async handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    this.setState({ loading: true })

    const amount = this.amount()
    if (amount === undefined) {
      return
    }
    const ok = await this.props.createChannel(this.state.counterparty, amount)

    if (ok) {
      this.props.closeModal()
      this.props.redirect && this.props.redirect(this.state.counterparty)
    } else {
      this.setState({ loading: false, showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }

  private formIsValid() {
    const amount = this.amount()
    return (
      amount &&
      amount > 0 &&
      validRecipientAccount(this.props.username, this.state.counterparty) &&
      this.walletHasSufficientBalance()
    )
  }

  private walletHasSufficientBalance() {
    const amount = this.amount()
    return amount && this.props.availableBalance >= amount
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    availableBalance: getWalletStroops(state),
    username: state.config.Username,
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    createChannel: (counterparty: string, initialDeposit: number) => {
      return createChannel(dispatch, counterparty, initialDeposit)
    },
  }
}

export const ConnectedCreateChannel = connect<
  {},
  {},
  {
    closeModal: () => void
    prefill?: { counterparty: string }
    redirect?: (account: string) => void
  }
>(
  mapStateToProps,
  mapDispatchToProps
)(CreateChannel)
