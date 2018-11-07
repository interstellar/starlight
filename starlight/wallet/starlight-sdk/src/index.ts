import {
  Update,
  ClientState,
  ClientResponse,
  UpdateHandler,
  ResponseHandler,
  Event,
} from './types'
import { getWalletOp, getChannelOps } from './helpers'

import { URL } from 'url'

require('isomorphic-fetch')

const StrKey = require('stellar-base').StrKey

export { ClientState, UpdateHandler, ResponseHandler, Update }

export const initialClientState: ClientState = {
  addressesForChannelAccount: {},
  addressesForCounterpartyAccount: {},
  updateNumberForSequenceNumber: {},
  from: 0,
}

export interface InitConfigParams {
  HorizonURL: string
  Password: string
  Username: string
}

interface EditParams {
  HorizonURL?: string
  OldPassword?: string
  Password?: string
}

/**
 * @typedef Status
 * @type {object}
 * @property {boolean} IsConfigured
 * @property {boolean} IsLoggedIn
 */

/**
 * @typedef ClientResponse
 * @template T
 * @type {object}
 * @property {string} body - The body of the response.
 * @property {boolean} ok - Whether the request was successful.
 * @property {number | undefined} status - Status code from the HTTP response (if any).
 * @property {error | undefined} error - a JavaScript error, if there was an error making the request.
 */

export interface ClientResponse {
  body: any
  ok: boolean
  status?: number
  error?: Error
}

/**
 * The Starlight API Client object is the root object for all API interactions.
 * To interact with Starlight, a Client object must always be instantiated
 * first.
 * @class
 */
export class Client {
  public updateHandler?: UpdateHandler
  public responseHandler?: ResponseHandler
  private cookie: string

  /**
   * Create a Client.
   * @param {string} baseURL - The URL of the Starlight agent.
   * @param {object} clientState - The client state.
   */
  constructor(
    public baseURL?: string,
    public clientState: ClientState = initialClientState
  ) {}

  /**
   * Resets the client's state to an initial state.
   */
  public async clearState() {
    this.clientState = initialClientState
  }

  /**
   * Restore the client's state from a snapshot.
   * @param {object} clientState - The client state.
   */
  public async setState(clientState: ClientState) {
    this.clientState = clientState
  }

  /**
   * Configure the instance with a username, password, and horizon URL.
   *
   * @param {object} params - The configuration parameters.
   * @param {string} params.HorizonURL - The Horizon URL (by default, https://horizon-testnet.stellar.org).
   * @param {string} params.Username - This will be the first part of your Stellar address (as in "alice*stellar.org").
   * @param {string} params.Password - This will also be used to encrypt the instance's private key in storage.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async configInit(params: InitConfigParams) {
    return this.request('/api/config-init', params)
  }

  /**
   * Edit the instance's configuration.
   * @param {object} params - The configuration parameters.
   * @param {string} [params.HorizonURL] - A new Horizon URL.
   * @param {string} [params.Password] - A new password.
   * @param {string} [params.OldPassword] - The old password, which must be provided if a new password is provided.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async configEdit(params: EditParams) {
    return this.request('/api/config-edit', params)
  }

  /**
   * Attempt to open a channel with a specific counterparty.
   * @param {string} counterpartyAddress - The Stellar address of your counterparty (e.g., "alice*stellar.org").
   * @param {number} initialDeposit - The amount (in stroops) you will initially deposit into the channel.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async createChannel(
    counterpartyAddress: string,
    initialDeposit: number
  ) {
    return this.request('/api/do-create-channel', {
      GuestAddr: counterpartyAddress,
      HostAmount: initialDeposit,
    })
  }

  /**
   * Cooperatively close a channel.
   * @param {string} channelID - The channel ID.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async close(channelID: string) {
    return this.request('/api/do-command', {
      ChannelID: channelID,
      Command: {
        Name: 'CloseChannel',
      },
    })
  }

  /**
   * Cancel a proposed channel that your counterparty has not yet accepted.
   * @param {string} channelID - The channel ID.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async cancel(channelID: string) {
    return this.request('/api/do-command', {
      ChannelID: channelID,
      Command: {
        Name: 'CleanUp',
      },
    })
  }

  /**
   * Attempt to force close a channel.
   * @param {string} channelID - The channel ID.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async forceClose(channelID: string) {
    return this.request('/api/do-command', {
      ChannelID: channelID,
      Command: {
        Name: 'ForceClose',
      },
    })
  }

  /**
   * Make a payment over a channel.
   * @param {string} channelID - The channel ID.
   * @param {number} amount - The amount (in stroops) to be paid.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async channelPay(channelID: string, amount: number) {
    return this.request('/api/do-command', {
      ChannelID: channelID,
      Command: {
        Name: 'ChannelPay',
        Amount: amount,
      },
    })
  }

  /**
   * Make a payment on the public network.
   * @param {string} channelID - The channel ID.
   * @param {number} amount - The amount (in stroops) to be paid.
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async walletPay(recipient: string, amount: number) {
    return this.request('/api/do-wallet-pay', {
      Dest: recipient,
      Amount: amount,
    })
  }

  /**
   * Add more money to a channel you created.
   * @param {string} channelID - The channel ID.
   * @param {number} amount - The amount (in stroops) to be deposited.
   * @returns {Promise<ClientReponse<string>>}
   */
  public async deposit(channelID: string, amount: number) {
    return this.request('/api/do-command', {
      ChannelID: channelID,
      Command: {
        Name: 'TopUp',
        Amount: amount,
      },
    })
  }

  /**
   * Authenticate with a Starlight instance.
   * This also decrypts the instance's private key,
   * allowing it to sign transactions and accept channels.
   * @param {string} username
   * @param {number} password
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async login(username: string, password: string) {
    return this.request('/api/login', {
      username,
      password,
    })
  }

  /**
   * Log out of a Starlight instance.
   * @param {string} username
   * @param {number} password
   *
   * @returns {Promise<ClientReponse<string>>}
   */
  public async logout() {
    return this.request('/api/logout')
  }

  /**
   * Find the account ID (e.g., "G...") corresponding to a Stellar address (e.g., "alice*stellar.org").
   * @param {string} address
   *
   * @returns {Promise<ClientResponse<status>} accountID
   */
  public async findAccount(address: string): Promise<ClientResponse> {
    return this.request('/api/find-account', {
      stellar_addr: address,
    })
  }

  /**
   * Get the current status of the instance (whether the instance is configured, and if so, whether the user is logged in).
   *
   * @returns {Promise<Status | undefined>}
   */
  public async getStatus() {
    return this.request('/api/status')
  }

  /**
   * Subscribe to updates from the Starlight instance.
   *
   * @param {function} updateHandler - A handler function that updates will be passed to.
   */
  public subscribe(updateHandler: UpdateHandler) {
    this.updateHandler = updateHandler
    this.loop()
  }

  /**
   * Stop subscribing to updates from the Starlight instance.
   */
  public unsubscribe() {
    this.updateHandler = undefined
  }

  /**
   * This method is only public so it can be used in tests.
   * @ignore
   */
  public handleFetchResponse(rawResponse: ClientResponse) {
    const handler = this.updateHandler

    const response = this.responseHandler
      ? this.responseHandler(rawResponse)
      : rawResponse

    if (handler && response.ok && response.body.length >= 1) {
      response.body.forEach((event: Event) => {
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
        this.clientState = {
          ...this.clientState,
          from: event.UpdateNum + 1,
        }
        const update = this.eventToUpdate(event)
        if (update) {
          handler(update)
        }
      })
    }
    return response.ok
  }

  private async fetch() {
    const From = this.clientState.from
    const response = await this.request('/api/updates', { From })

    return this.handleFetchResponse(response)
  }

  private async loop() {
    if (!this.updateHandler) {
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

  private eventToUpdate(event: Event): Update | undefined {
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

  private async request(path: string = '', data = {}) {
    let urlString: string
    if (this.baseURL) {
      const url = new URL(path, this.baseURL)
      urlString = url.href
    } else {
      urlString = path
    }
    const rawResponse = await this.post(urlString, data)
    const response = this.responseHandler
      ? this.responseHandler(rawResponse)
      : rawResponse
    return response
  }

  private async post(url = ``, data = {}): Promise<ClientResponse> {
    let response: Response
    try {
      // Default options marked with *
      response = await fetch(url, {
        method: 'POST',
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, same-origin, *omit
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
          Cookie: this.cookie,
        },
        body: JSON.stringify(data), // body data type must match "Content-Type"
      })

      const cookie = response.headers.get('set-cookie')
      if (cookie) {
        this.cookie = cookie
      }

      if ((response.headers.get('content-type') || '').includes('json')) {
        return {
          body: await response.json(),
          ok: response.ok,
          status: response.status,
        }
      } else {
        return { body: '', ok: response.ok, status: response.status }
      }
    } catch (error) {
      return { body: '', ok: false, error }
    }
  }
}
