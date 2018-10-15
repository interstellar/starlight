import * as React from 'react'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faCog,
  faCopy,
  faExchangeAlt,
  faInfoCircle,
  faWallet,
} from '@fortawesome/free-solid-svg-icons'
import { IconProp } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

// Explicitly import / store references to fa icons we are using
library.add(faCog, faCopy, faExchangeAlt, faInfoCircle, faWallet)

export class Icon extends React.Component<{
  name: IconProp
  className?: string
}> {
  public render() {
    return (
      <FontAwesomeIcon
        className={this.props.className}
        icon={this.props.name}
      />
    )
  }
}
