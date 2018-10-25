import * as React from 'react'
import styled from 'styled-components'

import { GraphSegment } from 'pages/shared/graphs/GraphSegment'
import { DUSTYGRAY } from 'pages/shared/Colors'
import { Tooltip } from 'pages/shared/Tooltip'

import { stroopsToLumens } from 'helpers/lumens'

const GraphWrapper = styled.span`
  display: flex;
  align-items: center;
`
const LabelWrapper = styled.span<{ align: string }>`
  margin-${props => props.align}: 10px;
  padding-bottom: 5px;
  text-align: ${props => props.align};
  width: auto;
`
const Label = styled.span`
  color: ${DUSTYGRAY};
  cursor: default;
  display: inline-block;
  font-family: 'Nitti Grotesk';
  font-size: 14px;
  font-weight: 500;
  text-transform: uppercase;
`
const SubLabel = styled.label<{ color: string }>`
  color: ${props => props.color};
  display: block;
  font-family: 'Nitti Grotesk';
  font-size: 18px;
  font-weight: 700;
  text-transform: uppercase;
`
const SegmentWrapper = styled.div`
  flex-grow: 2;
`

interface Props {
  leftLabel?: string
  leftTooltip?: string
  leftAmount: number
  rightAmount: number
  rightLabel?: string
  rightTooltip?: string
  leftColor: string
  rightColor: string
}

export class BarGraph extends React.Component<Props> {
  public constructor(props: Props) {
    super(props)
  }

  private totalAmount() {
    return this.props.leftAmount + this.props.rightAmount
  }

  private calculatePercentage(amount: number) {
    if (this.totalAmount() === 0) {
      return 0
    }
    return (amount / this.totalAmount()) * 100
  }

  public render() {
    return (
      <GraphWrapper>
        <LabelWrapper align="right">
          {this.props.leftTooltip ? (
            <Tooltip hover content={this.props.leftTooltip}>
              <Label>{this.props.leftLabel || 'Send'}</Label>
            </Tooltip>
          ) : (
            <Label>{this.props.leftLabel || 'Send'}</Label>
          )}
          <SubLabel color={this.props.leftColor}>
            {stroopsToLumens(this.props.leftAmount)} XLM
          </SubLabel>
        </LabelWrapper>
        <SegmentWrapper>
          <GraphSegment
            color={this.props.leftColor}
            height="25px"
            side="left"
            full={this.calculatePercentage(this.props.rightAmount) === 0}
            width={this.calculatePercentage(this.props.leftAmount)}
          />
          <GraphSegment
            color={this.props.rightColor}
            height="25px"
            side="right"
            full={this.calculatePercentage(this.props.leftAmount) === 0}
            width={this.calculatePercentage(this.props.rightAmount)}
          />
        </SegmentWrapper>
        <LabelWrapper align="left">
          {this.props.rightTooltip ? (
            <Tooltip hover content={this.props.rightTooltip}>
              <Label>{this.props.rightLabel || 'Receive'}</Label>
            </Tooltip>
          ) : (
            <Label>{this.props.rightLabel || 'Receive'}</Label>
          )}
          <SubLabel color={this.props.rightColor}>
            {stroopsToLumens(this.props.rightAmount)} XLM
          </SubLabel>
        </LabelWrapper>
      </GraphWrapper>
    )
  }
}
