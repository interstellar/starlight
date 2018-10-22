import * as React from 'react'
import styled from 'styled-components'
import { CopyToClipboard } from 'react-copy-to-clipboard'

import { CORNFLOWER } from 'pages/shared/Colors'
import { Icon } from 'pages/shared/Icon'
import { Tooltip } from 'pages/shared/Tooltip'

const CopyIconWrapper = styled.span`
  margin-left: 10px;
`

const CopyIcon = styled(Icon)`
  cursor: pointer;

  &:hover {
    color: ${CORNFLOWER};
  }
`

interface Props {
  id: string
  truncate?: boolean
}

export class CopyableString extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    return (
      <span>
        <span title={this.props.id}>
          {this.props.truncate
            ? this.truncateStr(this.props.id)
            : this.props.id}
        </span>
          <CopyToClipboard text={this.props.id} >
            <CopyIconWrapper>
              <Tooltip click content="Copied!">
                <CopyIcon name="copy" />
              </Tooltip>
            </CopyIconWrapper>
          </CopyToClipboard>
      </span>
    )
  }

  private truncateStr(str: string, length = 6) {
    if (!str) {
      return ''
    }

    const strLength = str.length

    return `${str.substring(0, length)}...${str.substring(strLength - length)}`
  }
}
