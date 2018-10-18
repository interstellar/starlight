import * as React from 'react'
import * as ReactTooltip from 'react-tooltip'
import styled from 'styled-components'

const DisabledBtnContainer = styled.div`
  display: inline-block;
`

interface Props {
  children: any
  content: string
  disable?: boolean
}

export class DisabledBtnHover extends React.Component<Props, {}> {
  public constructor(props: Props) {
    super(props)
  }

  public render() {
    return (
      <DisabledBtnContainer
        data-tip={this.props.content}
        data-tip-disable={this.props.disable}
      >
        <ReactTooltip effect="solid" multiline />

        { this.props.children }
      </DisabledBtnContainer>
    )
  }
}
