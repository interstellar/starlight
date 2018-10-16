import * as React from 'react'
import styled from 'styled-components'

import { RADICALRED, SEAFOAM } from 'components/styled/Colors'
import { Icon } from 'components/styled/Icon'

const Span = styled.span<{ isLoading?: boolean }>`
  color: ${props => (props.isLoading ? RADICALRED : SEAFOAM)};
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
      return <Span>{this.props.value}</Span>
    } else {
      return (
        <Span isLoading={true}>
          {this.props.value} <Icon className="fa-pulse" name="spinner" />
        </Span>
      )
    }
  }
}
