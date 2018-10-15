import * as React from 'react'
import styled from 'styled-components'

import { CORNFLOWER, DUSTYGRAY } from 'components/styled/Colors'

const Container = styled.span`
  margin-top: 14px;
`
const Radio = styled.input`
  -moz-appearance: none;
  -webkit-appearance: none;
  appearance: none;
  border-radius: 50%;
  border: 1px solid ${DUSTYGRAY};
  height: 18px;
  outline: none;
  position: relative;
  top: 6px;
  transition: border 0.2s ease-in-out;
  width: 18px;

  &:checked {
    border: 6px solid ${CORNFLOWER};
  }
`
const Label = styled.label`
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-style: normal;
  font-weight: normal;
  margin-left: 11px;
  margin-right: 35px;
`

export class RadioButton extends React.Component<{
  name: string
  text: string
  checked?: boolean
  onClick?: (event: React.MouseEvent<any>) => void
}> {
  public render() {
    const id = this.props.text.replace(' ', '-')

    return (
      <Container>
        <Radio
          type="radio"
          name={this.props.name}
          id={id}
          defaultChecked={this.props.checked}
          onClick={this.props.onClick}
        />
        <Label htmlFor={id}>{this.props.text}</Label>
      </Container>
    )
  }
}
