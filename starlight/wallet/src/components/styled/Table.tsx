import styled from 'styled-components'

import { ALTO, EBONYCLAY } from 'components/styled/Colors'

export const Table = styled.table`
  border-collapse: collapse;
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 500;
  text-align: left;
  width: 100%;
`
export const TableHeaderRow = styled.tr`
  border-bottom: 1px solid ${ALTO};
  font-size: 14px;
`
export const TableHeader = styled.th<{ align: string }>`
  padding: 15px 0;
  text-align: ${props => props.align};
`
export const TableData = styled.td<{ align: string; color?: string }>`
  color: ${props => props.color || EBONYCLAY};
  padding: 15px 0;
  text-align: ${props => props.align};
`
