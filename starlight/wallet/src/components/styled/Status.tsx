import styled from 'styled-components'

import { RADICALRED, SEAFOAM } from 'components/styled/Colors'

export const Status = styled.span<{ value: string }>`
  color: ${props => (props.value === 'Open' ? SEAFOAM : RADICALRED)};
`
