import * as React from 'react'
import styled from 'styled-components'
import { CopyToClipboard } from 'react-copy-to-clipboard'

import { CORNFLOWER } from 'pages/shared/Colors'
import { Icon } from 'pages/shared/Icon'

const CopyIcon = styled(Icon)`
  cursor: pointer;
  margin-left: 10px;

  &:hover {
    color: ${CORNFLOWER};
  }
`
const CopiedLabel = styled.span<{ copied: boolean }>`
  color: ${CORNFLOWER};
  margin-left: 10px;
  opacity: ${props => (props.copied ? '1' : '0')};
  transition: opacity 0.25s;
`

interface Props {
  id: string
  truncate?: boolean
}

interface State {
  copied: boolean
  timer?: number
}

export class CopyableString extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      copied: false,
      timer: undefined,
    }
  }

  public componentWillUnmount() {
    clearTimeout(this.state.timer)
  }

  public render() {
    return (
      <span>
        <span title={this.props.id}>
          { this.props.truncate ?
            this.truncateStr(this.props.id) :
            this.props.id
          }
        </span>
        <CopyToClipboard
          text={this.props.id}
          onCopy={() => {
            clearTimeout(this.state.timer)

            this.setState({
              copied: true,
              timer: window.setTimeout(() => {
                this.setState({ copied: false })
              }, 3000),
            })
          }}
        >
          <span>
            <CopyIcon name="copy" />
          </span>
        </CopyToClipboard>
        <CopiedLabel copied={this.state.copied}>Copied!</CopiedLabel>
      </span>
    )
  }

  private truncateStr(str: string, length = 6) {
    if (!str) { return "" }

    const strLength = str.length

    return `${str.substring(0, length)}...${str.substring(
      strLength - length,
    )}`
  }
}
