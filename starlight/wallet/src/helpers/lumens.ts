import Big from 'big.js'

export const stroopsToLumens = (stroops: number, options: any = {}) => {
  return formatAmount(stroops * 0.0000001, options.short && 2)
}

export const lumensToStroops = (lumens: number) => {
  return Math.floor(
    Number(
      Big(lumens).times(Big(10000000))
    )
  )
}

export const formatAmount = (amount: number, maxDigits = 7) => {
  return amount.toLocaleString(undefined, { maximumFractionDigits: maxDigits })
}
