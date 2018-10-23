import * as moment from 'moment'

export function fromNowPast(timestamp: string) {
  // handles clock drift case where timestamp is slightly in the future
  const now = moment()
  return moment(moment(timestamp) < now ? timestamp : now).fromNow()
}
