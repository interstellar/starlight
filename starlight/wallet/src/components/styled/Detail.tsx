import styled from 'styled-components'

import { DUSTYGRAY } from 'components/styled/Colors'

export const Detail = styled.div`
  margin-bottom: 15px;

  &:last-child {
    margin-bottom: 0;
  }
`
export const DetailLabel = styled.span`
  color: ${DUSTYGRAY};
  display: inline-block;
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  font-weight: 500;
  width: 130px;
`
export const DetailValue = styled.span`
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 500;
  width: 470px;
`
