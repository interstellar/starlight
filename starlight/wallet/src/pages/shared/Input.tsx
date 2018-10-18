import styled from 'styled-components'

import { ALTO, CORNFLOWER, DUSTYGRAY } from 'pages/shared/Colors'

export const Label = styled.label`
  display: inline-block;
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 10px;
  text-transform: uppercase;
`
export const Hint = styled.span`
  color: ${DUSTYGRAY};
  float: right;
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  margin-top: 5px;
`
export const Input = styled.input`
  border-radius: 5px;
  border: 1px solid ${ALTO};
  box-sizing: border-box;
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 500;
  margin-bottom: 45px;
  outline: none;
  padding: 18px;
  transition: border 0.2s ease-in-out;
  width: 100%;

  ::-webkit-inner-spin-button,
  ::-webkit-outer-spin-button {
    appearance: none;
    margin: 0;
  }

  &:focus {
    border: 1px solid ${CORNFLOWER};
  }
`

export const HelpBlock = styled.div<{ isShowing: boolean }>`
  border-radius: 5px;
  border: solid ${CORNFLOWER} 1px;
  color: ${CORNFLOWER};
  display: ${props => (props.isShowing ? 'block' : 'none')};
  font-family: 'Nitti Grotesk';
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 45px;
  padding: 20px;
`
