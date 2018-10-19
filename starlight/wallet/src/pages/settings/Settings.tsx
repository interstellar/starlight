import * as React from 'react'
import { Dispatch } from 'redux'
import { connect } from 'react-redux'

import { ApplicationState } from 'types/schema'
import { BtnHeading } from 'pages/shared/Button'
import { ConnectedChangePassword } from 'pages/settings/ChangePassword'
import { ConnectedChangeServer } from 'pages/settings/ChangeServer'
import { Container } from 'pages/shared/Container'
import { Detail, DetailLabel, DetailValue } from 'pages/shared/Detail'
import { Heading } from 'pages/shared/Heading'
import { Link } from 'pages/shared/Link'
import { Modal } from 'pages/shared/Modal'
import { Section, SectionHeading } from 'pages/shared/Section'
import { RADICALRED } from 'pages/shared/Colors'
import { lifecycle } from 'state/lifecycle'

interface Props {
  Username: string
  HorizonURL: string
  logout: () => any
}
interface State {
  openedModalName: string
  showFlash: boolean
  flashMessage: string
}

export class Settings extends React.Component<Props, State> {
  public constructor(props: any) {
    super(props)

    this.state = {
      openedModalName: '',
      showFlash: false,
      flashMessage: '',
    }

    this.openModal = this.openModal.bind(this)
    this.closeModal = this.closeModal.bind(this)
    this.showFlash = this.showFlash.bind(this)
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

  private showFlash(message: string) {
    this.setState({ showFlash: true, flashMessage: message })

    window.setTimeout(() => {
      this.setState({ showFlash: false, flashMessage: '' })
    }, 3000)
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
            <DetailValue>Testnet</DetailValue>
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
                <ConnectedChangeServer closeModal={() => this.closeModal()} />
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
                <ConnectedChangePassword closeModal={() => this.closeModal()} />
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
