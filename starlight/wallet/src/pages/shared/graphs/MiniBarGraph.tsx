import * as React from 'react'
import styled from 'styled-components'

import { GraphSegment } from 'pages/shared/graphs/GraphSegment'
import { ALTO, CORNFLOWER, EBONYCLAY } from 'pages/shared/Colors'

import { formatAmount, stroopsToLumens } from 'helpers/lumens'

const GraphWrapper = styled.span`
  align-items: center;
  display: flex;
`
const Label = styled.label<{ color: string }>`
  color: ${props => props.color};
  display: inline-block;
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 700;
  margin: 10px;
  text-transform: uppercase;
  white-space: nowrap;
`
const SegmentWrapper = styled.span<{ position: string }>`
  align-items: center;
  display: flex;
  flex: 1;
  justify-content: ${props => props.position};
`
const VerticalLineWrapper = styled.span`
  flex: 0 0 auto;
  justify-content: center;
  line-height: 0;
`
const VerticalLine = styled.div`
  background: ${ALTO};
  display: inline-block;
  height: 50px;
  width: 2px;
`

interface Props {
  leftAmount: number
  rightAmount: number
}

export class MiniBarGraph extends React.Component<Props> {
  public constructor(props: Props) {
    super(props)
  }

  private totalAmount = this.props.leftAmount + this.props.rightAmount

  private calculatePercentage(amount: number) {
    if (this.totalAmount === 0) {
      return 0
    }
    return (amount / this.totalAmount) * 100
  }

  public render() {
    return (
      <GraphWrapper>
        <SegmentWrapper position="flex-end">
          <Label color={CORNFLOWER}>
            {formatAmount(
              stroopsToLumens(this.props.leftAmount, { short: true })
            )}{' '}
            XLM
          </Label>
          <GraphSegment
            color={CORNFLOWER}
            height="10px"
            side="left"
            width={this.calculatePercentage(this.props.leftAmount)}
          />
        </SegmentWrapper>
        <VerticalLineWrapper>
          <VerticalLine />
        </VerticalLineWrapper>
        <SegmentWrapper position="flex-start">
          <GraphSegment
            color={EBONYCLAY}
            height="10px"
            side="right"
            width={this.calculatePercentage(this.props.rightAmount)}
          />
          <Label color={EBONYCLAY}>
            {formatAmount(
              stroopsToLumens(this.props.rightAmount, { short: true })
            )}{' '}
            XLM
          </Label>
        </SegmentWrapper>
      </GraphWrapper>
    )
  }
}
