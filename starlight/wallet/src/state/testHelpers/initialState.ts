import { ApplicationState } from 'types/schema'
import { initialClientState } from 'starlight-sdk'

export const initialState: ApplicationState = {
  lifecycle: {
    isLoggedIn: false,
    isConfigured: false,
  },
  config: {
    Username: '',
    HorizonURL: 'https://horizon-testnet.stellar.org',
  },
  events: {
    clientState: initialClientState,
  },
  wallet: {
    ID: '',
    Balance: 0,
    Ops: [],
    Pending: {},
    AccountAddresses: {},
  },
  channels: {},
  flash: {
    message: '',
    showFlash: false,
  },
}
