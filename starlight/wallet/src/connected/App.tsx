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
import { Navigation } from 'Navigation'
import { lifecycle } from 'state/lifecycle'
import { hot } from 'react-hot-loader'

interface Props {
  isConfigured: boolean
  isLoggedIn: boolean
  status: () => any
}

class App extends React.Component<Props, {}> {
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

              <Route path="/channel/*" exact={true} component={Navigation} />

              <Route path="/settings" exact={true} component={Navigation} />

              <Route path="/" exact={true} component={Navigation} />

              <Route path="/*" render={() => <Redirect to="/wallet" />} />
            </Switch>
          </ConnectedEventLoop>
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
          <Route path="/*" render={() => <Redirect to="/" />} />}
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
          <Route path="/*" render={() => <Redirect to="/" />} />}
        </Switch>
      )
    }
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    isConfigured: state.lifecycle.isConfigured,
    isLoggedIn: state.lifecycle.isLoggedIn,
  }
}
const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    status: () => lifecycle.status(dispatch),
  }
}
export const ConnectedApp = connect(
  mapStateToProps,
  mapDispatchToProps
)(hot(module)(App))
