import * as numeral from 'numeral'

export const stroopsToLumens = (stroops: number, options: any = {}) => {
  const lumens = numeral(stroops * 0.0000001)
  if (options.short) {
    return lumens.format('0[.]00')
  } else {
    return lumens.format('0[.]00[00000]')
  }
}

export const lumensToStroops = (lumens: number) => {
  return Math.floor(lumens * 10000000)
}

export const formatAmount = (amount: string) => {
  return Number(amount).toLocaleString()
}
