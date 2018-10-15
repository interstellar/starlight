import * as React from 'react'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'
import styled from 'styled-components'

import { BtnSubmit } from 'components/styled/Button'
import { RADICALRED } from 'components/styled/Colors'
import { Heading } from 'components/styled/Heading'
import { Input, Label } from 'components/styled/Input'
import { config } from 'state/config'

const View = styled.div`
  padding: 25px;
`
const Form = styled.form`
  margin-top: 45px;
`

interface Props {
  editPassword: (params: { OldPassword: string; NewPassword: string }) => any
}

interface State {
  OldPassword: string
  NewPassword: string
  showError: boolean
}

export class ChangePassword extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      OldPassword: '',
      NewPassword: '',
      showError: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    return (
      <View>
        <Heading>Change password</Heading>
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="OldPassword">Current Password</Label>
          <Input
            type="password"
            name="OldPassword"
            autoComplete="off"
            autoFocus
            onChange={e => {
              this.setState({ OldPassword: e.target.value })
            }}
          />
          <Label htmlFor="NewPassword">New Password</Label>
          <Input
            type="password"
            name="NewPassword"
            autoComplete="off"
            onChange={e => {
              this.setState({ NewPassword: e.target.value })
            }}
          />

          {this.formatSubmitButton()}
        </Form>
      </View>
    )
  }

  private formatSubmitButton() {
    if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Error changing password
        </BtnSubmit>
      )
    } else {
      return <BtnSubmit>Save</BtnSubmit>
    }
  }

  private async handleSubmit() {
    const ok = await this.props.editPassword({
      OldPassword: this.state.OldPassword,
      NewPassword: this.state.NewPassword,
    })

    if (!ok) {
      this.setState({ showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    editPassword: (params: { OldPassword: string; NewPassword: string }) => {
      return config.edit(dispatch, params)
    },
  }
}

export const ConnectedChangePassword = connect(
  null,
  mapDispatchToProps
)(ChangePassword)
