import * as React from 'react'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import styled from 'styled-components'

import { validRecipientAccount } from 'helpers/account'
import { stroopsToLumens } from 'helpers/lumens'

import { BtnSubmit } from 'pages/shared/Button'
import { Heading } from 'pages/shared/Heading'
import { Icon } from 'pages/shared/Icon'
import { Hint, Input, Label } from 'pages/shared/Input'
import { HorizontalLine } from 'pages/shared/HorizontalLine'
import { Total } from 'pages/shared/Total'
import { Unit, UnitContainer } from 'pages/shared/Unit'
import { CORNFLOWER, RADICALRED } from 'pages/shared/Colors'
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
  AvailableBalance: number
  closeModal: () => void
  createChannel: (recipient: string, initialDeposit: number) => void
  prefill?: { counterparty: string }
  redirect?: (account: string) => void
  username: string
}

interface State {
  Counterparty: string
  InitialDeposit: string
  showError: boolean
  loading: boolean
  formErrors: {
    deposit: boolean,
    counterparty: boolean
  }
}

export class CreateChannel extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      Counterparty: this.props.prefill ? this.props.prefill.counterparty : '',
      InitialDeposit: '',
      showError: false,
      loading: false,
      formErrors: {
        deposit: false,
        counterparty: false,
      },
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    return (
      <View>
        <Heading>Create channel</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="Counterparty">Counterparty</Label>
          <Input
            value={this.state.Counterparty}
            onBlur={() => {
              if (
                this.state.Counterparty &&
                !validRecipientAccount(
                  this.props.username,
                  this.state.Counterparty
                )
              ) {
                this.setState({ formErrors: {
                  deposit: this.state.formErrors.deposit,
                  counterparty: true,
                }})
              } else {
                this.setState({ formErrors: {
                  deposit: this.state.formErrors.deposit,
                  counterparty: false,
                },
              })
            }}}
            onChange={e => {
              this.setState({ Counterparty: e.target.value })
            }}
            type="text"
            name="Counterparty"
            autoComplete="off"
            autoFocus={!this.state.Counterparty}
            error={this.state.formErrors.counterparty}
          />

          <Label htmlFor="InitialDeposit">Initial Deposit</Label>
          <Hint>
            <strong>{stroopsToLumens(this.props.AvailableBalance)} XLM</strong>{' '}
            available in account
          </Hint>
          <UnitContainer>
            <Input
              value={this.state.InitialDeposit}
              onBlur={() => {
                if (this.state.InitialDeposit && !this.walletHasSufficientBalance()) {
                  this.setState({ formErrors: {
                    deposit: true,
                    counterparty: this.state.formErrors.counterparty,
                  }})
                } else {
                  this.setState({ formErrors: {
                    deposit: false,
                    counterparty: this.state.formErrors.counterparty,
                  },
                })
              }}}
              onChange={e => {
                this.setState({ InitialDeposit: e.target.value })
              }}
              type="number"
              name="InitialDeposit"
              autoComplete="off"
              autoFocus={!!this.state.Counterparty}
              error={this.state.formErrors.deposit}
            />
            <Unit>XLM</Unit>
          </UnitContainer>

          <HalfWidth>
            <Label>Transaction Fee</Label>
            <Amount>0.00001 XLM</Amount>
          </HalfWidth>

          <HalfWidth>
            <Label>
              Channel Reserve <InfoIcon name="info-circle" />
            </Label>
            <Amount>5 XLM</Amount>
          </HalfWidth>

          <HorizontalLine />

          <Label>Total Required</Label>
          <Total>&mdash;</Total>

          {this.formatSubmitButton()}
        </Form>
      </View>
    )
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

    const ok = await this.props.createChannel(
      this.state.Counterparty,
      parseInt(this.state.InitialDeposit, 10) * 10000000
    )

    if (ok) {
      this.props.closeModal()
      this.props.redirect && this.props.redirect(this.state.Counterparty)
    } else {
      this.setState({ loading: false, showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }

  private formIsValid() {
    return (
      validRecipientAccount(this.props.username, this.state.Counterparty) &&
      this.walletHasSufficientBalance()
    )
  }

  private walletHasSufficientBalance() {
    return this.props.AvailableBalance >=
      (parseFloat(this.state.InitialDeposit) * 10000000)
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    AvailableBalance: getWalletStroops(state),
    username: state.config.Username,
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    createChannel: (Counterparty: string, InitialDeposit: number) => {
      return createChannel(dispatch, Counterparty, InitialDeposit)
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
