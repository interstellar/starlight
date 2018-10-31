import { Event } from 'types/types'
import { Update, ClientState } from 'client/types'
import { getWalletOp, getChannelOps } from 'client/helpers'

const StrKey = require('stellar-base').StrKey

type Handler = (event: Update) => void

export async function post(url = ``, data = {}) {
  let response
  try {
    // Default options marked with *
    response = await fetch(url, {
      method: 'POST',
      cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
      credentials: 'same-origin', // include, same-origin, *omit
      headers: {
        'Content-Type': 'application/json; charset=utf-8',
      },
      body: JSON.stringify(data), // body data type must match "Content-Type"
    })

    if (response.status === 401) {
      return { body: await response.json(), ok: response.ok, loggedIn: false }
    }

    if ((response.headers.get('content-type') || '').includes('json')) {
      return { body: await response.json(), ok: response.ok, loggedIn: true }
    } else {
      return { body: '', ok: response.ok, loggedIn: true }
    }
  } catch (error) {
    return { body: '', ok: true, loggedIn: true }
  }
}

export const initialClientState: ClientState = {
  addressesForChannelAccount: {},
  addressesForCounterpartyAccount: {},
  updateNumberForSequenceNumber: {},
  from: 0,
}

export class Client {
  public handler?: Handler

  constructor(
    public clientState: ClientState = initialClientState,
    private onLogout: () => void
  ) {}

  public async fetch() {
    const From = this.clientState.from
    const response = await post('/api/updates', { From })

    return this.handleResponse(response)
  }

  public handleResponse(response: {
    body: any
    ok: boolean
    loggedIn: boolean
  }) {
    if (!response.loggedIn) {
      return this.onLogout()
    }

    const handler = this.handler

    if (handler && response.ok && response.body.length >= 1) {
      response.body.forEach((event: Event) => {
        this.clientState = {
          ...this.clientState,
          from: event.UpdateNum + 1,
        }
        this.updateState(event)
        const update = this.eventToUpdate(event)
        if (update) {
          handler(update)
        }
      })
    }
    return response.ok
  }

  public subscribe(handler: Handler) {
    this.handler = handler
    this.loop()
  }

  public updateState(event: Event) {
    if (event.Type === 'channel') {
      const channel = event.Channel
      const CounterpartyAccount =
        channel.Role === 'Host' ? channel.GuestAcct : channel.HostAcct
      this.clientState = {
        ...this.clientState,
        addressesForCounterpartyAccount: {
          ...this.clientState.addressesForCounterpartyAccount,
          [CounterpartyAccount]: channel.CounterpartyAddress,
        },
        addressesForChannelAccount: {
          ...this.clientState.addressesForChannelAccount,
          [channel.EscrowAcct]: channel.CounterpartyAddress,
        },
      }
    }
  }

  public unsubscribe() {
    this.handler = undefined
  }

  private async loop() {
    if (!this.handler) {
      return
    }
    const ok = await this.fetch()
    if (!ok) {
      await this.backoff(10000)
    }

    this.loop()
  }

  private async backoff(ms: number) {
    await new Promise(resolve => setTimeout(resolve, ms))
  }

  public eventToUpdate(event: Event): Update | undefined {
    const clientState = this.clientState
    switch (event.Type) {
      case 'account': {
        const op = getWalletOp(
          event,
          this.clientState.addressesForCounterpartyAccount
        )
        if (event.InputTx) {
          // check if this is from a channel account
          const counterpartyAddress = this.clientState
            .addressesForChannelAccount[
            StrKey.encodeEd25519PublicKey(
              event.InputTx.Env.Tx.SourceAccount.Ed25519
            )
          ]
          if (counterpartyAddress !== undefined) {
            // it's from a channel
            // activity is handled elsewhere
            return {
              Type: 'accountUpdate',
              Account: event.Account,
              UpdateLedgerTime: event.UpdateLedgerTime,
              UpdateNum: event.UpdateNum,
              ClientState: clientState,
            }
          }
        }
        return {
          Type: 'walletActivityUpdate',
          Account: event.Account,
          UpdateLedgerTime: event.UpdateLedgerTime,
          UpdateNum: event.UpdateNum,
          WalletOp: op,
          ClientState: clientState,
        }
      }
      case 'channel':
        // TODO: remove channel account from this mapping when channel is closed
        const ops = getChannelOps(event)
        if (ops.length === 0) {
          return {
            Type: 'channelUpdate',
            Account: event.Account,
            Channel: event.Channel,
            UpdateLedgerTime: event.UpdateLedgerTime,
            UpdateNum: event.UpdateNum,
            ClientState: clientState,
          }
        } else {
          return {
            Type: 'channelActivityUpdate',
            Account: event.Account,
            Channel: event.Channel,
            UpdateLedgerTime: event.UpdateLedgerTime,
            UpdateNum: event.UpdateNum,
            ChannelOp: ops[0],
            ClientState: clientState,
          }
        }
      case 'config':
        return {
          Type: 'configUpdate',
          Config: event.Config,
          Account: event.Account,
          UpdateLedgerTime: event.UpdateLedgerTime,
          UpdateNum: event.UpdateNum,
          ClientState: clientState,
        }
      case 'init':
        return {
          Type: 'initUpdate',
          Config: event.Config,
          Account: event.Account,
          UpdateLedgerTime: event.UpdateLedgerTime,
          UpdateNum: event.UpdateNum,
          ClientState: clientState,
        }
      case 'tx_failed':
      case 'tx_success':
        return {
          Type:
            event.Type === 'tx_failed' ? 'txFailureUpdate' : 'txSuccessUpdate',
          Tx: event.InputTx,
          Account: event.Account,
          UpdateLedgerTime: event.UpdateLedgerTime,
          UpdateNum: event.UpdateNum,
          ClientState: clientState,
        }
    }
  }
}
