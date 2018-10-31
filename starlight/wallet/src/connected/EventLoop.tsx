import * as React from 'react'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'

import { ApplicationState } from 'types/schema'
import { events } from 'state/events'
import { Client } from 'client/client'
import { Update, ClientState } from 'client/types'
import { LOGOUT_SUCCESS } from 'state/lifecycle'

interface Props {
  handler: (update: Update) => any
  clientState: ClientState
  dispatchLogout: () => void
}

class EventLoop extends React.Component<Props, {}> {
  private client: Client

  public async componentDidMount() {
    this.client = new Client(this.props.clientState, this.props.dispatchLogout)
    this.client.subscribe(this.props.handler)
  }

  public render() {
    return this.props.children
  }

  public async componentWillUnmount() {
    this.client.unsubscribe()
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    clientState: state.events.clientState,
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    handler: events.getHandler(dispatch),
    dispatchLogout: () => {
      dispatch({
        type: LOGOUT_SUCCESS,
      })
    },
  }
}

export const ConnectedEventLoop = connect<{}, {}, {}>(
  mapStateToProps,
  mapDispatchToProps
)(EventLoop)
