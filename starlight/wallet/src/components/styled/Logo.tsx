import * as React from 'react'
import styled from 'styled-components'

import { CORNFLOWER } from 'components/styled/Colors'

const Container = styled.div`
  text-align: center;
`
const Wordmark = styled.span`
  color: white;
  display: block;
  font-family: 'Avenir Next', sans-serif;
  font-size: 24px;
  text-transform: uppercase;
`

export const Logo = () => {
  return (
    <Container>
      <svg width="50px" height="50px" viewBox="0 0 50 50">
        <g
          id="Welcome"
          stroke="none"
          strokeWidth="1"
          fill="none"
          fillRule="evenodd"
        >
          <g
            id="Login"
            transform="translate(-676.000000, -92.000000)"
            fill={CORNFLOWER}
          >
            <g id="Group-2" transform="translate(400.000000, 92.000000)">
              <g id="Group" transform="translate(237.000000, 0.000000)">
                <polygon
                  id="Star"
                  points="63.7266403 30.0485188 39 49.4532805 58.4047617 24.7266403 39 -1.0125234e-13 63.7266403 19.4047617 88.4532805 -1.08357767e-13 69.0485188 24.7266403 88.4532805 49.4532805"
                />
              </g>
            </g>
          </g>
        </g>
      </svg>
      <Wordmark>Starlight</Wordmark>
    </Container>
  )
}
