import * as React from 'react'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'

import { ApplicationState } from 'schema'
import { events } from 'state/events'

interface Props {
  fetch: (From: number) => any
  From: number
}

class EventLoop extends React.Component<Props, {}> {
  private stop: boolean

  public async componentDidMount() {
    this.stop = false
    this.loop()
  }

  public render() {
    return this.props.children
  }

  private async loop() {
    if (this.stop) {
      return
    }
    const ok = await this.tick()
    if (!ok) {
      await this.backoff(10000)
    }

    this.loop()
  }

  private async tick() {
    return await this.props.fetch(this.props.From)
  }

  public async componentWillUnmount() {
    this.stop = true
  }

  private async backoff(ms: number) {
    await new Promise(resolve => setTimeout(resolve, ms))
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    From: state.events.From,
  }
}
const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    fetch: (From: number) => events.fetch(dispatch, From),
  }
}
export const ConnectedEventLoop = connect(
  mapStateToProps,
  mapDispatchToProps
)(EventLoop)
