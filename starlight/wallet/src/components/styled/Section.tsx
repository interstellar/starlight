import styled from 'styled-components'

import { CORNFLOWER, WHITE } from 'components/styled/Colors'

export const Section = styled.div`
  background: ${WHITE};
  border-radius: 5px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-top: 45px;
  padding: 30px 40px;
`

export const SectionHeading = styled.label`
  color: ${CORNFLOWER};
  display: block;
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 25px;
  text-transform: uppercase;
`
