import * as React from 'react'
import { connect } from 'react-redux'

import { ApplicationState } from 'schema'

import {
  ConnectedWalletActivityTable
} from 'components/wallet/WalletActivityTable'
import { ConnectedWalletBalance } from 'components/wallet/WalletBalance'
import { BtnHeading } from 'components/styled/Button'
import { Container } from 'components/styled/Container'
import { CopyableString } from 'components/styled/CopyableString'
import { Detail, DetailLabel, DetailValue } from 'components/styled/Detail'
import { Heading } from 'components/styled/Heading'
import { Modal } from 'components/styled/Modal'
import { Section, SectionHeading } from 'components/styled/Section'
import { ConnectedSendPayment } from 'components/forms/SendPayment'

const StrKey = require('stellar-base').StrKey

interface Props {
  id: string
  username: string
}

interface State {
  hasOpenModal: boolean
}

export class Wallet extends React.Component<Props, State> {
  public constructor(props: Props) {
    super(props)

    this.state = {
      hasOpenModal: false,
    }

    this.openModal = this.openModal.bind(this)
    this.closeModal = this.closeModal.bind(this)
  }

  private openModal() {
    this.setState({ hasOpenModal: true })
  }

  private closeModal() {
    this.setState({ hasOpenModal: false })
  }

  public componentDidMount() {
    document.title = `Wallet - ${this.props.username}`
  }

  public render() {
    const address = `${this.props.username}*${window.location.host}`
    return (
      <Container>
        <Heading>Wallet</Heading>
        <BtnHeading onClick={this.openModal}>Send</BtnHeading>
        <Modal isOpen={this.state.hasOpenModal} onClose={this.closeModal}>
          <ConnectedSendPayment closeModal={() => this.closeModal()} />
        </Modal>
        <Section>
          <SectionHeading>Balance</SectionHeading>
          <ConnectedWalletBalance />
        </Section>
        <Section>
          <SectionHeading>Account Details</SectionHeading>
          <Detail>
            <DetailLabel>Address</DetailLabel>
            <DetailValue>
              <CopyableString
                id={address}
                truncate={StrKey.isValidEd25519PublicKey(address)}
              />
            </DetailValue>
          </Detail>
          <Detail>
            <DetailLabel>Account ID </DetailLabel>
            <DetailValue>
              <CopyableString
                id={this.props.id}
                truncate={StrKey.isValidEd25519PublicKey(this.props.id)}
              />
            </DetailValue>
          </Detail>
        </Section>
        <Section>
          <SectionHeading>Activity</SectionHeading>
          <ConnectedWalletActivityTable />
        </Section>
      </Container>
    )
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    id: state.wallet.ID,
    username: state.config.Username,
  }
}
export const ConnectedWallet = connect(
  mapStateToProps,
  {}
)(Wallet)
