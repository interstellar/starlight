import * as React from 'react'
import styled from 'styled-components'

import { RADICALRED, SEAFOAM } from 'pages/shared/Colors'
import { Icon } from 'pages/shared/Icon'

const Span = styled.span<{ color: string }>`
  color: ${props => props.color};
`

interface Props {
  value: string
}

export class Status extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    if (this.props.value === 'Open') {
      return <Span color={SEAFOAM}>{this.props.value}</Span>
    } else if (this.props.value === 'Closed') {
      return <Span color={RADICALRED}>{this.props.value}</Span>
    } else {
      return (
        <Span color={RADICALRED}>
          {this.props.value} <Icon className="fa-pulse" name="spinner" />
        </Span>
      )
    }
  }
}
