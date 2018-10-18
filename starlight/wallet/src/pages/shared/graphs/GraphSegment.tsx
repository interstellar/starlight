import * as React from 'react'
import styled from 'styled-components'

const Segment = styled.div<{
  width: number
  color: string
  height: string
  side: string
  full?: boolean
}>`
  background: ${props => props.color};
  ${props => props.full ?
    `border-radius: ${props.height}` :
    `
      border-top-${props.side}-radius: ${props.height};
      border-bottom-${props.side}-radius: ${props.height};
    `}
  display: inline-block;
  height: ${props => props.height};
  width: ${props => props.width}%;
`

interface Props {
  color: string
  height: string
  side: string
  width: number
  full?: boolean
}

export class GraphSegment extends React.Component<Props> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    return (
      <Segment
        color={this.props.color}
        height={this.props.height}
        side={this.props.side}
        width={this.props.width}
        full={this.props.full}
      />
    )
  }
}
