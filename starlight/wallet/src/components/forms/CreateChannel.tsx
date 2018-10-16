import * as React from 'react'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import styled from 'styled-components'

import { BtnSubmit } from 'components/styled/Button'
import { Heading } from 'components/styled/Heading'
import { Hint, Input, Label } from 'components/styled/Input'
import { Icon } from 'components/styled/Icon'
import { HorizontalLine } from 'components/styled/HorizontalLine'
import { Total } from 'components/styled/Total'
import { Unit, UnitContainer } from 'components/styled/Unit'
import { CORNFLOWER, RADICALRED } from 'components/styled/Colors'
import { ApplicationState } from 'schema'
import { getWalletStroops } from 'state/wallet'
import { createChannel } from 'state/channels'
import { stroopsToLumens } from 'lumens'

const View = styled.div`
  padding: 25px;
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
const Amount = styled.div`
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 45px;
  text-transform: uppercase;
`

interface Props {
  AvailableBalance: number
  closeModal: () => void
  createChannel: (recipient: string, initialDeposit: number) => void
  prefill?: { counterparty: string }
  redirect?: (account: string) => void
}

interface State {
  Counterparty: string
  InitialDeposit: string // TODO(croaky): number?
  showError: boolean
}

export class CreateChannel extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      Counterparty: this.props.prefill ? this.props.prefill.counterparty : '',
      InitialDeposit: '',
      showError: false,
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
            onChange={e => {
              this.setState({ Counterparty: e.target.value })
            }}
            type="text"
            name="Counterparty"
            autoComplete="off"
            autoFocus={!this.state.Counterparty}
          />

          <Label htmlFor="InitialDeposit">Initial Deposit</Label>
          <Hint>
            <strong>{stroopsToLumens(this.props.AvailableBalance)} XLM</strong>{' '}
            available in account
          </Hint>
          <UnitContainer>
            <Input
              value={this.state.InitialDeposit}
              onChange={e => {
                this.setState({ InitialDeposit: e.target.value })
              }}
              type="number"
              name="InitialDeposit"
              autoComplete="off"
              autoFocus={!!this.state.Counterparty}
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
    if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Error opening channel
        </BtnSubmit>
      )
    } else {
      return <BtnSubmit>Open channel</BtnSubmit>
    }
  }

  private async handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()

    const ok = await this.props.createChannel(
      this.state.Counterparty,
      parseInt(this.state.InitialDeposit, 10) * 10000000
    )

    if (ok) {
      this.props.closeModal()
      this.props.redirect && this.props.redirect(this.state.Counterparty)
    } else {
      this.setState({ showError: true })
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
    createChannel: (Counterparty: string, InitialDeposit: number) => {
      return createChannel(dispatch, Counterparty, InitialDeposit)
    },
  }
}

export const ConnectedCreateChannel = connect<
  {},
  {},
  {
    closeModal: () => void;
    prefill?: { counterparty: string };
    redirect?: (account: string) => void;
  }
>(
  mapStateToProps,
  mapDispatchToProps
)(CreateChannel)
