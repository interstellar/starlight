import * as React from 'react'
import * as renderer from 'react-test-renderer'

import { ChangePassword } from 'pages/settings/ChangePassword'
import { ChangeServer } from 'pages/settings/ChangeServer'
import { CreateChannel } from 'pages/shared/forms/CreateChannel'
import { Deposit } from 'pages/channel/Deposit'
import { SendPayment } from 'pages/shared/forms/SendPayment'

const editPasswordFunc = (params: { OldPassword: string; Password: string }) =>
  params
const editServerFunc = (params: { HorizonURL: string }) => params
const createChannel = (_1: string, _2: number) => undefined
const showFlash = () => undefined

const sendFunc = async (_1: string, _2: number) => {
  return undefined
}
const closeModal = () => undefined
const deposit = (_1: string, _2: number) => undefined

it('renders ChangePassword', () => {
  const tree = renderer
    .create(
      <ChangePassword
        closeModal={closeModal}
        showFlash={showFlash}
        editPassword={editPasswordFunc}
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders ChangeServer', () => {
  const tree = renderer
    .create(
      <ChangeServer
        HorizonURL="https://horizon-testnet.stellar.org"
        editServer={editServerFunc}
        closeModal={closeModal}
        showFlash={showFlash}
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders CreateChannel', () => {
  const tree = renderer
    .create(
      <CreateChannel
        AvailableBalance={0}
        closeModal={closeModal}
        createChannel={createChannel}
        username="jessicard"
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders Deposit', () => {
  const tree = renderer
    .create(
      <Deposit
        channel={{} as any}
        deposit={deposit}
        AvailableBalance={500}
        closeModal={closeModal}
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders SendPayment', () => {
  const tree = renderer
    .create(
      <SendPayment
        AvailableBalance={1000}
        walletPay={sendFunc}
        channelPay={sendFunc}
        closeModal={closeModal}
        Channels={{}}
        CounterpartyAccounts={{}}
        username="jessicard"
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})
