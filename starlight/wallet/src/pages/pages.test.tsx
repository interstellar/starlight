import * as React from 'react'
import { Provider } from 'react-redux'
import { MemoryRouter } from 'react-router'
import * as renderer from 'react-test-renderer'
import configureMockStore from 'redux-mock-store'

const mockStore = configureMockStore()

import { ConfigLanding } from 'pages/config/ConfigLanding'
import { Channels } from 'pages/channels/Channels'
import { Credentials } from 'types/types'
import { InitConfig } from 'pages/config/InitConfig'
import { LoginForm } from 'pages/login/LoginForm'
import { Login } from 'pages/login/Login'
import { Settings } from 'pages/settings/Settings'
import { Wallet } from 'pages/wallet/Wallet'

const configFunc = (params: {
  Username: string
  Password: string
  HorizonURL: string
}) => params

const loginFunc = (params: Credentials) => params

it('renders Config', () => {
  const tree = renderer
    .create(<ConfigLanding form={<InitConfig configure={configFunc} />} />)
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders Login', () => {
  const tree = renderer
    .create(<Login form={<LoginForm login={loginFunc} />} />)
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders Channels', () => {
  const tree = renderer
    .create(
      <MemoryRouter>
        <Channels
          channels={[]}
          location={{}}
          totalChannelBalance={50000}
          totalChannelCounterpartyBalance={30000}
          username="alice"
        />
      </MemoryRouter>
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders Settings', () => {
  const tree = renderer
    .create(
      <Settings
        Username="croaky"
        HorizonURL="https://horizon-testnet.stellar.org"
        logout={() => null}
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders Wallet', () => {
  const store = mockStore({
    channels: {},
    wallet: { Ops: [] },
  })

  const tree = renderer
    .create(
      <Provider store={store}>
        <Wallet
          username="alice"
          id="GDQEYK27FM4LZCV54D7XB75DR76BGJJYJEKNGREPAVARTYA27KHL6H32"
        />
      </Provider>
    )
    .toJSON()
  expect(tree).toBeTruthy()
})
