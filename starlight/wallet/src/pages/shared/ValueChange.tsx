import * as React from 'react'
import styled from 'styled-components'

import { DUSTYGRAY, RADICALRED, SEAFOAM } from 'pages/shared/Colors'
import { stroopsToLumens } from 'helpers/lumens'

const Container = styled.span<{ color: string }>`
  color: ${props => props.color};
`

export class ValueChange extends React.Component<{ value: number }, {}> {
  public render() {
    if (this.props.value === 0) {
      return <Container color={DUSTYGRAY}>&mdash;</Container>
    } else if (this.props.value > 0) {
      return (
        <Container color={SEAFOAM}>
          + {stroopsToLumens(this.props.value)} XLM
        </Container>
      )
    } else {
      return (
        <Container color={RADICALRED}>
          - {stroopsToLumens(Math.abs(this.props.value))} XLM
        </Container>
      )
    }
  }
}
