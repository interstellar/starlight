import * as React from 'react'
import * as ReactModal from 'react-modal'

const style = {
  overlay: {
    backgroundColor: 'rgba(0, 0, 0, 0.75)',
    overflow: 'scroll',
  },
  content: {
    border: 'none',
    bottom: 'auto',
    margin: '45px auto',
    width: '600px',
  },
}

interface Props {
  children: any
  isOpen: boolean
  onClose: () => void
}

export class Modal extends React.Component<Props, {}> {
  public render() {
    return (
      <ReactModal
        isOpen={this.props.isOpen}
        onRequestClose={() => this.props.onClose()}
        style={style}
        ariaHideApp={false}
      >
        {this.props.children}
      </ReactModal>
    )
  }
}
