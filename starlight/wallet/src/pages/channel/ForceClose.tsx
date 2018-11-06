import * as React from 'react'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import styled from 'styled-components'

import { BtnSubmit } from 'pages/shared/Button'
import { Heading } from 'pages/shared/Heading'

import { forceClose, getWithdrawalTime } from 'state/channels'
import { ChannelState } from 'types/schema'

const Form = styled.form`
  margin-top: 45px;
`
const View = styled.div`
  padding: 25px;
`
const ConfirmationMessage = styled.p`
  margin: 0 0 45px 0;
`

interface Props {
  closeModal: () => void
  forceClose: (id: string) => Promise<boolean>
  channel: ChannelState
}

interface State {
  showError: boolean
  loading: boolean
}

export class ForceClose extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      showError: false,
      loading: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    return (
      <View>
        <Heading>Force close</Heading>
        <Form onSubmit={this.handleSubmit}>
          <ConfirmationMessage>
            Are you sure you want to force close the channel? If you do, your
            funds will be available after:
            <br />
            <br />
            <strong>{getWithdrawalTime(this.props.channel)}</strong>
          </ConfirmationMessage>
          <BtnSubmit>Force close</BtnSubmit>
        </Form>
      </View>
    )
  }

  private async handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    this.setState({ loading: true })

    const ok = await this.props.forceClose(this.props.channel.ID)

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

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    forceClose: (id: string) => {
      return forceClose(dispatch, id)
    },
  }
}

export const ConnectedForceClose = connect<
  {},
  {},
  {
    channel: ChannelState
    closeModal: () => void
  }
>(
  null,
  mapDispatchToProps
)(ForceClose)
