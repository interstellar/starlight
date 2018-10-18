import * as React from 'react'
import styled from 'styled-components'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'

import { Arrow } from 'pages/shared/Arrow'
import { Heading } from 'pages/shared/Heading'
import { Hint, Input, Label } from 'pages/shared/Input'
import { Icon } from 'pages/shared/Icon'
import { RadioButton } from 'pages/shared/RadioButton'
import { BtnSubmit } from 'pages/shared/Button'
import { ConfigState } from 'types/schema'
import { RADICALRED, WHITE } from 'pages/shared/Colors'
import { config } from 'state/config'

const View = styled.div`
  background: ${WHITE};
  border-radius: 5px;
  margin-top: 45px;
  padding: 45px;
`
const Form = styled.form`
  margin-top: 45px;
`
const RadioGroup = styled.div`
  margin-bottom: 45px;
`

interface Props {
  configure: (params: ConfigState) => any
}

interface State {
  Password: string
  DemoServer: boolean
  showError: boolean
  loading: boolean
}

export class InitConfig extends React.Component<Props, ConfigState & State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      Username: '',
      Password: '',
      DemoServer: true,
      HorizonURL: 'https://horizon-testnet.stellar.org',
      showError: false,
      loading: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    return (
      <View>
        <Heading>Configure your instance</Heading>{' '}
        <Form onSubmit={this.handleSubmit}>
          <div>
            <Label htmlFor="Username">Username</Label>
            <Hint>
              Wallet address: {this.state.Username || 'username'}*
              {window.location.host}
            </Hint>
          </div>
          <Input
            value={this.state.Username}
            onChange={e => {
              this.setState({ Username: e.target.value })
            }}
            type="text"
            name="Username"
            autoComplete="off"
            autoFocus
          />

          <Label htmlFor="Password">Password</Label>
          <Input
            value={this.state.Password}
            onChange={e => {
              this.setState({ Password: e.target.value })
            }}
            type="password"
            name="Password"
          />

          <RadioGroup>
            <div>
              <Label htmlFor="Testnet">Network</Label>
            </div>
            <RadioButton name="Testnet" text="Testnet" checked />
          </RadioGroup>

          <div>
            <RadioGroup>
              <div>
                <Label htmlFor="HorizonURL">Horizon API Server</Label>
              </div>
              <RadioButton
                name="HorizonURLChooser"
                text="Use demo server"
                checked={this.state.DemoServer}
                onClick={() => {
                  this.setState({
                    DemoServer: true,
                    HorizonURL: 'https://horizon-testnet.stellar.org',
                  })
                }}
              />
              <RadioButton
                name="HorizonURLChooser"
                text="Provide server URL"
                checked={!this.state.DemoServer}
                onClick={() => {
                  this.setState({
                    DemoServer: false,
                    HorizonURL: 'https://horizon-testnet.stellar.org',
                  })
                }}
              />
            </RadioGroup>
            {!this.state.DemoServer && (
              <div>
                <Label htmlFor="HorizonURL">Server URL</Label>
                <Input
                  type="text"
                  name="HorizonURL"
                  autoComplete="off"
                  autoFocus
                  onChange={e => {
                    this.setState({ HorizonURL: e.target.value })
                  }}
                />
              </div>
            )}
          </div>

          {this.formatSubmitButton()}
        </Form>
      </View>
    )
  }

  private formatSubmitButton() {
    if (this.state.loading) {
      return (
        <BtnSubmit disabled>
          Configuring <Icon className="fa-pulse" name="spinner" />
        </BtnSubmit>
      )
    } else if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Error configuring
        </BtnSubmit>
      )
    } else {
      return (
        <BtnSubmit>
          Configure <Arrow>&rarr;</Arrow>
        </BtnSubmit>
      )
    }
  }

  private async handleSubmit(event: any) {
    event.preventDefault()
    this.setState({ loading: true })

    const ok = await this.props.configure(this.state)

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
    configure: (params: ConfigState) => {
      return config.init(dispatch, params)
    },
  }
}
export const ConnectedInitConfig = connect(
  null,
  mapDispatchToProps
)(InitConfig)
