import * as React from 'react'
import { connect } from 'react-redux'
import { Dispatch } from 'redux'
import styled from 'styled-components'

import { ApplicationState } from 'schema'
import { RADICALRED } from 'components/styled/Colors'
import { Heading } from 'components/styled/Heading'
import { Input, Label } from 'components/styled/Input'
import { RadioButton } from 'components/styled/RadioButton'
import { BtnSubmit } from 'components/styled/Button'
import { config } from 'state/config'

const View = styled.div`
  padding: 25px;
`
const Form = styled.form`
  margin-top: 45px;
`
const RadioGroup = styled.div`
  margin-bottom: 45px;
`

interface Props {
  HorizonURL: string
  editServer: (params: { HorizonURL: string }) => any
}
interface State {
  DemoServer: boolean
  HorizonURL: string
  showError: boolean
}

export class ChangeServer extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      HorizonURL: this.props.HorizonURL,
      DemoServer:
        this.props.HorizonURL === 'https://horizon-testnet.stellar.org',
      showError: false,
    }

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  public render() {
    return (
      <View>
        <Heading>Horizon API Server</Heading>
        <Form onSubmit={this.handleSubmit}>
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
                    HorizonURL: '',
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
    if (this.state.showError) {
      return (
        <BtnSubmit color={RADICALRED} disabled>
          Error changing server
        </BtnSubmit>
      )
    } else {
      return <BtnSubmit>Save</BtnSubmit>
    }
  }

  private async handleSubmit() {
    const ok = await this.props.editServer({
      HorizonURL: this.state.HorizonURL,
    })

    if (!ok) {
      this.setState({ showError: true })
      window.setTimeout(() => {
        this.setState({ showError: false })
      }, 3000)
    }
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    HorizonURL: state.config.HorizonURL,
  }
}
const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    editServer: (params: { HorizonURL: string }) => {
      return config.edit(dispatch, params)
    },
  }
}
export const ConnectedChangeServer = connect(
  mapStateToProps,
  mapDispatchToProps
)(ChangeServer)
