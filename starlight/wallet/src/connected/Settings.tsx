import * as React from 'react'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'

import { ApplicationState } from 'schema'
import { BtnHeading } from 'components/styled/Button'
import { ConnectedChangePassword } from 'components/forms/ChangePassword'
import { ConnectedChangeServer } from 'components/forms/ChangeServer'
import { Container } from 'components/styled/Container'
import { Detail, DetailLabel, DetailValue } from 'components/styled/Detail'
import { Heading } from 'components/styled/Heading'
import { Link } from 'components/styled/Link'
import { Modal } from 'components/styled/Modal'
import { Section, SectionHeading } from 'components/styled/Section'
import { RADICALRED } from 'components/styled/Colors'
import { lifecycle } from 'state/lifecycle'

export class Settings extends React.Component<
  {
    Username: string
    HorizonURL: string
    logout: () => any
  },
  {
    openedModalName: string
  }
> {
  public constructor(props: any) {
    super(props)

    this.state = {
      openedModalName: '',
    }

    this.openModal = this.openModal.bind(this)
    this.closeModal = this.closeModal.bind(this)
  }

  private openModal(name: string) {
    this.setState({ openedModalName: name })
  }

  private hasOpenModal(name: string) {
    return this.state.openedModalName === name
  }

  private closeModal() {
    this.setState({ openedModalName: '' })
  }

  public componentDidMount() {
    document.title = `Settings - ${this.props.Username}`
  }

  public render() {
    return (
      <Container>
        <Heading>Settings</Heading>
        <BtnHeading color={RADICALRED} onClick={this.props.logout}>
          Log Out
        </BtnHeading>
        <Section>
          <SectionHeading>Configuration</SectionHeading>
          <Detail>
            <DetailLabel>Network</DetailLabel>
            <DetailValue>
              Testnet
            </DetailValue>
          </Detail>
          <Detail>
            <DetailLabel>Horizon API Server</DetailLabel>
            <DetailValue>
              {this.props.HorizonURL ? this.props.HorizonURL : 'Demo server'}{' '}
              <Link onClick={() => this.openModal('server')}>(edit)</Link>
              <Modal
                isOpen={this.hasOpenModal('server')}
                onClose={this.closeModal}
              >
                <ConnectedChangeServer />
              </Modal>
            </DetailValue>
          </Detail>
        </Section>
        <Section>
          <SectionHeading>User Details</SectionHeading>
          <Detail>
            <DetailLabel>Username</DetailLabel>
            <DetailValue>{this.props.Username} </DetailValue>
          </Detail>
          <Detail>
            <DetailLabel>Password</DetailLabel>
            <DetailValue>
              <Link onClick={() => this.openModal('password')}>
                Change Password
              </Link>
              <Modal
                isOpen={this.hasOpenModal('password')}
                onClose={this.closeModal}
              >
                <ConnectedChangePassword />
              </Modal>
            </DetailValue>
          </Detail>
        </Section>
      </Container>
    )
  }
}

const mapStateToProps = (state: ApplicationState) => {
  return {
    Username: state.config.Username,
    HorizonURL: state.config.HorizonURL,
  }
}
const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    logout: () => lifecycle.logout(dispatch),
  }
}
export const ConnectedSettings = connect(
  mapStateToProps,
  mapDispatchToProps
)(Settings)
