import styled from 'styled-components'

import { RADICALRED, SEAFOAM, WILDSAND } from 'pages/shared/Colors'

export const Flash = styled.div`
  align-items: center;
  background-color: ${SEAFOAM};
  border-radius: 5px;
  color: ${WILDSAND};
  display: flex;
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 500;
  height: 70px;
  justify-content: center;
  left: 200px;
  margin-left: auto;
  margin-right: auto;
  position: absolute;
  right: 0;
  top: 30px;
  width: 400px;
`

export const AlertFlash = styled(Flash)`
  background-color: ${RADICALRED};
`
