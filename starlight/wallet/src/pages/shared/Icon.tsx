import * as React from 'react'
import * as ReactTooltip from 'react-tooltip'
import styled from 'styled-components'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faCog,
  faCopy,
  faExchangeAlt,
  faInfoCircle,
  faSpinner,
  faWallet,
} from '@fortawesome/free-solid-svg-icons'
import { IconProp } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

// Explicitly import / store references to fa icons we are using
library.add(faCog, faCopy, faExchangeAlt, faInfoCircle, faSpinner, faWallet)

const Span = styled.span`
  text-transform: none;
`

interface Props {
  name: IconProp
  className?: string
  tooltipContent?: string
}

export class Icon extends React.Component<Props, {}> {
  public render() {
    return (
      <Span>
        <ReactTooltip effect="solid" multiline />
        <FontAwesomeIcon
          className={this.props.className}
          data-tip={this.props.tooltipContent}
          icon={this.props.name}
        />
      </Span>
    )
  }
}
