import * as React from 'react'
import { Dispatch } from 'redux'
import { Route, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { connect } from 'react-redux'

import { ApplicationState } from 'types/schema'
import { ConfigLanding } from 'pages/config/ConfigLanding'
import { ConnectedEventLoop } from 'connected/EventLoop'
import { ConnectedInitConfig } from 'pages/config/InitConfig'
import { ConnectedLoginForm } from 'pages/login/LoginForm'
import { Login } from 'pages/login/Login'
import { Flash } from 'pages/shared/Flash'
import { Navigation } from 'Navigation'
import { lifecycle } from 'state/lifecycle'
import { hot } from 'react-hot-loader'

import { flash } from 'state/flash'
import { RADICALRED } from 'pages/shared/Colors'

interface Props {
  isConfigured: boolean
  isLoggedIn: boolean
  flash: {
    message: string
    color: string
    show: boolean
  }
  location: any
  status: () => any
  setFlash: (message: string, color: string) => void
  clearFlash: () => void
}

class App extends React.Component<Props, {}> {
  public constructor(props: any) {
    super(props)

    // on page reload
    if (performance && performance.navigation.type === 1) {
      this.props.clearFlash()
    }
  }

  public async componentDidMount() {
    this.props.status()
  }

  public render() {
    if (this.props.isLoggedIn) {
      return (
        <div>
          <ConnectedEventLoop>
            <Switch>
              <Route path="/wallet" exact={true} component={Navigation} />

              <Route path="/channels" exact={true} component={Navigation} />

              <Route
                path="/channel/:address+"
                exact={true}
                component={Navigation}
              />

              <Route path="/settings" exact={true} component={Navigation} />

              <Route path="/" exact={true} component={Navigation} />

              <Route
                path="/*"
                render={() => {
                  this.props.setFlash('That page does not exist', RADICALRED)
                  return <Redirect to="/wallet" />
                }}
              />
            </Switch>
          </ConnectedEventLoop>
          {this.props.flash.show && (
            <Flash color={this.props.flash.color}>
              {this.props.flash.message}
            </Flash>
          )}
        </div>
      )
    } else if (this.props.isConfigured) {
      return (
        <Switch>
          <Route
            path="/"
            exact={true}
            render={props => <Login {...props} form={<ConnectedLoginForm />} />}
          />
          <Route
            path="/*"
            render={() => {
              this.props.setFlash('That page does not exist', RADICALRED)
              return <Redirect to="/" />
            }}
          />
        </Switch>
      )
    } else {
      return (
        <Switch>
          <Route
            path="/"
            exact={true}
            render={props => (
              <ConfigLanding {...props} form={<ConnectedInitConfig />} />
            )}
          />
          <Route
            path="/*"
            render={() => {
              this.props.setFlash('That page does not exist', RADICALRED)
              return <Redirect to="/" />
            }}
          />
        </Switch>
      )
    }
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    isConfigured: state.lifecycle.isConfigured,
    isLoggedIn: state.lifecycle.isLoggedIn,
    flash: {
      message: state.flash.message,
      color: state.flash.color,
      show: state.flash.showFlash,
    },
  }
}
const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    status: () => lifecycle.status(dispatch),
    setFlash: (message: string, color: string) => {
      return flash.set(dispatch, message, color)
    },
    clearFlash: () => {
      return flash.clear(dispatch)
    },
  }
}
export const ConnectedApp = connect<{}, {}, {}>(
  mapStateToProps,
  mapDispatchToProps
)(hot(module)(App))
