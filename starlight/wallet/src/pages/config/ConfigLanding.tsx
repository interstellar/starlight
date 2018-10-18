import * as React from 'react'

import { LandingContainer } from 'pages/shared/LandingContainer'
import { Logo } from 'pages/shared/Logo'

export class ConfigLanding extends React.Component<{ form: any }, {}> {
  public render() {
    return (
      <LandingContainer>
        <Logo />
        {this.props.form}
      </LandingContainer>
    )
  }
}
