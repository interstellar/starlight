import * as React from 'react'
import styled from 'styled-components'

import { LandingContainer } from 'pages/shared/LandingContainer'
import { Logo } from 'pages/shared/Logo'

const LogoContainer = styled.div`
  margin-bottom: 80px;
`

export class ConfigLanding extends React.Component<{ form: any }, {}> {
  public render() {
    return (
      <LandingContainer>
        <LogoContainer>
          <Logo large={true} />
        </LogoContainer>
        {this.props.form}
      </LandingContainer>
    )
  }
}
