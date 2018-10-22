import * as React from 'react'
import styled from 'styled-components'

import { EBONYCLAY, WHITE } from 'pages/shared/Colors'

const TooltipContainer = styled.span<{ show: boolean, direction: string }>`
  ${props => (props.direction === 'top' && 'bottom: 150%')};
  ${props => (props.direction === 'bottom' && 'top: 150%')};

  border-radius: 4px;
  color: ${WHITE};
  background: ${EBONYCLAY};
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  font-weight: bold;
  left: 50%;
  visibility: ${props => (props.show ? 'visible' : 'hidden')};
  opacity: ${props => (props.show ? '1' : '0')};
  padding: 5px 10px;
  position: absolute;
  text-align: center;
  text-transform: none;
  transform: translateX(-50%);
  transition: all 0.25s;
  white-space: nowrap;

  &::before {
    border: 5px solid transparent;
    ${props => (props.direction === 'top' &&
      `border-top-color: ${EBONYCLAY};
       top: 100%;`
    )}
    ${props => (props.direction === 'bottom' &&
      `border-bottom-color: ${EBONYCLAY};
       bottom: 100%;`
    )}
    content: ' ';
    height: 0;
    left: calc(50% - 5px);
    position: absolute;
    width: 0;
  }
`
const Wrapper = styled.span`
  display: inline-block;
  position: relative;
`

interface State {
  show: boolean
  timer?: number
}

interface Props {
  direction?: string
  children: any
  content: string
  click?: boolean
  hover?: boolean
}

export class Tooltip extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      show: false,
      timer: undefined,
    }
  }

  public componentWillUnmount() {
    clearTimeout(this.state.timer)
  }

  private onClick() {
    if (this.props.click) {
      clearTimeout(this.state.timer)

      this.setState({
        show: true,
        timer: window.setTimeout(() => {
          this.setState({ show: false })
        }, 3000)})
    }
  }

  private onMouseEnter() {
    this.props.hover && this.setState({ show: true })
  }

  private onMouseLeave() {
    this.props.hover && this.setState({ show: false })
  }

  public render() {
    return (
      <Wrapper
        onClick={() => this.onClick()}
        onMouseEnter={() => this.onMouseEnter()}
        onMouseLeave={() => this.onMouseLeave()}
      >
        <TooltipContainer
          direction={this.props.direction || "top"}
          dangerouslySetInnerHTML={{ __html: this.props.content }}
          show={this.state.show}
        />

        { this.props.children }
      </Wrapper>
    )
  }
}
