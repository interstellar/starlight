import * as React from 'react'
import styled, { injectGlobal } from 'styled-components'
import { Redirect, Route } from 'react-router'
import { Link, NavLink as ReactNavLink } from 'react-router-dom'

import { ConnectedChannel } from 'pages/channel/Channel'
import { ConnectedChannels } from 'pages/channels/Channels'
import { ConnectedSettings } from 'pages/settings/Settings'
import { ConnectedWallet } from 'pages/wallet/Wallet'
import { Icon } from 'pages/shared/Icon'
import { Interstellar } from 'pages/shared/Interstellar'
import { Logo } from 'pages/shared/Logo'
import {
  CORNFLOWER,
  CORNFLOWER_DARK,
  EBONYCLAY,
  WHITE,
  WILDSAND,
} from 'pages/shared/Colors'

const Container = styled.div`
  background: ${WILDSAND};
  display: flex;
  min-width: 920px;
`
const Footer = styled.div`
  flex-shrink: 0;
  margin: 20px auto;
`
const Links = styled.div`
  flex: 1 0 auto;
`
const LogoLink = styled(Link)`
  background: ${CORNFLOWER_DARK};
  display: block;
  padding: 45px 0;
  text-decoration: none;
`
const Nav = styled.div`
  background: ${EBONYCLAY};
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  position: fixed;
  width: 200px;
`
const NavLink = styled(ReactNavLink)`
  color: white;
  display: block;
  font-size: 18px;
  margin: 20px 0;
  padding: 0 40px;
  text-decoration: none;

  &:hover {
    color: ${CORNFLOWER};
  }
`
const NavIcon = styled(Icon)`
  margin-right: 10px;
`
const Preview = styled.span`
  color: ${WHITE};
  display: inline-block;
  font-size: 12px;
  text-align: center;
  width: 100%;
`
const View = styled.div`
  background: ${WILDSAND};
  flex: 1;
  margin-left: 200px;
  min-height: 100vh;
`

export class Navigation extends React.Component<{}, {}> {
  public globals: any

  public componentDidMount() {
    this.globals = injectGlobal`
      body {
        background: ${WILDSAND};
      }
    `
  }

  public componentWillUnmount() {
    this.globals = injectGlobal`
      body {
        background: ${EBONYCLAY};
      }
    `
  }

  public render() {
    return (
      <Container>
        <Nav>
          <Links>
            <LogoLink to="/">
              <Logo />
              <Preview>developer preview</Preview>
            </LogoLink>
            <NavLink to="/wallet" activeStyle={{ color: CORNFLOWER }}>
              <NavIcon name="wallet" />
              Wallet
            </NavLink>
            <NavLink to="/channels" activeStyle={{ color: CORNFLOWER }}>
              <NavIcon name="exchange-alt" />
              Channels
            </NavLink>
            <NavLink to="/settings" activeStyle={{ color: CORNFLOWER }}>
              <NavIcon name="cog" />
              Settings
            </NavLink>
          </Links>
          <Footer>
            <Interstellar />
          </Footer>
        </Nav>

        <View>
          <Route exact path="/" render={() => <Redirect to="/wallet" />} />
          <Route
            exact={true}
            path="/wallet"
            render={() => <ConnectedWallet />}
          />
          <Route exact={true} path="/channels" component={ConnectedChannels} />
          <Route path="/channel/:id" component={ConnectedChannel} />
          <Route exact={true} path="/settings" component={ConnectedSettings} />
        </View>
      </Container>
    )
  }
}
