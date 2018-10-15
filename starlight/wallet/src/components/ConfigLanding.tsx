import * as React from 'react'

import { LandingContainer } from 'components/styled/LandingContainer'
import { Logo } from 'components/styled/Logo'

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
