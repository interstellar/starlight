import * as numeral from 'numeral'
import Big from 'big.js'

export const stroopsToLumens = (stroops: number, options: any = {}) => {
  const singleStroop = 0.0000001
  const lumens = stroops * singleStroop

  if (lumens <= singleStroop) {
    return Big(lumens).toFixed()
  }

  if (options.short) {
    return numeral(lumens).format('0[.]00')
  } else {
    return numeral(lumens).format('0[.]00[00000]')
  }
}

export const lumensToStroops = (lumens: number) => {
  return Math.floor(
    Number(
      Big(lumens).times(Big(10000000))
    )
  )
}

export const formatAmount = (amount: string) => {
  return Number(amount).toLocaleString()
}
