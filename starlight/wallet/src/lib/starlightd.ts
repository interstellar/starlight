import { Client } from 'starlight-sdk'

// this client should only be used for making requests, not for processing updates
// it does not have any state
const client = new Client()

export const Starlightd = {
  client,
}
