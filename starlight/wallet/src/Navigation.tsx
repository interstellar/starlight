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
  WILDSAND,
} from 'pages/shared/Colors'

const NAV_WIDTH = '243px;'

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
  padding: 43px 40px 30px;
  text-decoration: none;
`
const Nav = styled.div`
  background: ${EBONYCLAY};
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  position: fixed;
  width: ${NAV_WIDTH};
`
const NavLink = styled(ReactNavLink)`
  color: white;
  display: flex;
  align-items: center;
  font-size: 18px;
  margin: 29px 0;
  padding: 0 40px;
  text-decoration: none;


  &:hover {
    color: ${CORNFLOWER};
  }
`
const NavIcon = styled(Icon)`
  margin-right: 15px;
`
const View = styled.div`
  background: ${WILDSAND};
  flex: 1;
  margin-left: ${NAV_WIDTH};
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
