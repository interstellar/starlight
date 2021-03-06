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
const createChannel = async (_1: string, _2: number) => true
const setFlash = () => undefined

const sendFunc = async (_1: string, _2: number) => true
const closeModal = () => undefined
const deposit = async (_1: string, _2: number) => true

it('renders ChangePassword', () => {
  const tree = renderer
    .create(
      <ChangePassword
        closeModal={closeModal}
        setFlash={setFlash}
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
        setFlash={setFlash}
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})

it('renders CreateChannel', () => {
  const tree = renderer
    .create(
      <CreateChannel
        availableBalance={0}
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
        availableBalance={500}
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
        availableBalance={1000}
        walletPay={sendFunc}
        channelPay={sendFunc}
        closeModal={closeModal}
        channels={{}}
        counterpartyAccounts={{}}
        username="jessicard"
      />
    )
    .toJSON()
  expect(tree).toBeTruthy()
})
