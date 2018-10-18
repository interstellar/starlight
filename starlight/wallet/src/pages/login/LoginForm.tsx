import * as React from 'react'
import styled from 'styled-components'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'

import { Arrow } from 'pages/shared/Arrow'
import { Credentials } from 'types/types'
import { Heading } from 'pages/shared/Heading'
import { Icon } from 'pages/shared/Icon'
import { Input, Label } from 'pages/shared/Input'
import { BtnSubmit } from 'pages/shared/Button'
import { CORNFLOWER, RADICALRED, WHITE } from 'pages/shared/Colors'
import { lifecycle } from 'state/lifecycle'

const View = styled.div`
  background: ${WHITE};
  border-radius: 5px;
  margin-top: 45px;
  padding: 45px;
`
const Form = styled.form`
  margin-top: 45px;
`

interface State {
  Username: string
  Password: string
  showError: boolean
  loading: boolean
}

export class LoginForm extends React.Component<
  { login: (params: Credentials) => any },
  State
> {
  public constructor(props: any) {
    super(props)

    this.state = {
      Username: '',
      Password: '',
      showError: false,
      loading: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    return (
      <View>
        <Heading>Login to your instance</Heading>{' '}
        <Form onSubmit={this.handleSubmit}>
          <Label htmlFor="Username">Username</Label>
          <Input
            value={this.state.Username}
            onChange={e => {
              this.setState({ Username: e.target.value })
            }}
            type="text"
            name="Username"
            autoComplete="off"
            autoFocus
            required
          />

          <Label htmlFor="Password">Password</Label>
          <Input
            value={this.state.Password}
            onChange={e => {
              this.setState({ Password: e.target.value })
            }}
            type="password"
            name="Password"
            required
          />

          {this.formatSubmitButton()}
        </Form>
      </View>
    )
  }

  private formatSubmitButton() {
    if (this.state.loading) {
      return (
        <BtnSubmit disabled>
          Logging in <Icon className="fa-pulse" name="spinner" />
        </BtnSubmit>
      )
    } else if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Invalid username or password
        </BtnSubmit>
      )
    } else {
      return (
        <BtnSubmit color={CORNFLOWER}>
          Login <Arrow>&rarr;</Arrow>
        </BtnSubmit>
      )
    }
  }

  private async handleSubmit(event: any) {
    event.preventDefault()
    this.setState({ loading: true })

    const ok = await this.props.login(this.state)

    if (!ok) {
      this.setState({ loading: false, showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    login: (params: Credentials) => {
      return lifecycle.login(dispatch, params)
    },
  }
}
export const ConnectedLoginForm = connect(
  null,
  mapDispatchToProps
)(LoginForm)
