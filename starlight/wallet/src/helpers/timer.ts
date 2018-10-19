export const timer = (callback: any, time: number) => {
  setTimeout(() => {
    callback()
  }, time)
}
