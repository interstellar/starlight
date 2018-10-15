import * as React from 'react'
import { render } from 'react-dom'
import { Provider } from 'react-redux'
import { Route } from 'react-router'
import { PersistGate } from 'redux-persist/integration/react'
import { ConnectedRouter } from 'connected-react-router'

import { ConnectedApp } from 'connected/App'
import { history, persistor, store } from 'lib/store'

// Global, non-styled component styles
// - Font family
// - Root body element
import 'assets/global.css'

// Set favicon
const favicon = document.createElement('link')
favicon.type = 'image/png'
favicon.rel = 'shortcut icon'
favicon.href = require('!!file?name=favicon.ico!assets/images/favicon.png')
document.getElementsByTagName('head')[0].appendChild(favicon)
document.title = 'Starlight'

// Start app
render(
  <Provider store={store}>
    <PersistGate loading={null} persistor={persistor}>
      <ConnectedRouter history={history}>
        <Route path="/" component={ConnectedApp} />
      </ConnectedRouter>
    </PersistGate>
  </Provider>,
  document.getElementById('root')
)
