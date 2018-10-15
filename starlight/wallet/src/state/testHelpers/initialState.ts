import { ApplicationState } from 'schema'

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
    From: 1,
    list: [],
  },
  wallet: {
    ID: '',
    Balance: 0,
    Ops: [],
    Pending: {},
  },
  channels: {},
}
