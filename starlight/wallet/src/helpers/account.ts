const StrKey = require('stellar-base').StrKey

export const validPublicKey = (account: string) => {
  return StrKey.isValidEd25519PublicKey(account)
}

export const validAddress = (address: string) => {
  return address.includes('*')
}

export const validAccount = (account: string) => {
  return validAddress(account) || validPublicKey(account)
}

export const usernameToAddress = (username: string) => {
  return `${username}*${window.location.host}`
}

export const validRecipientAccount = (
  currentUsername: string,
  recipient: string
) => {
  if (recipient === usernameToAddress(currentUsername)) {
    return false
  } else {
    return validAccount(recipient)
  }
}
