import { createStore, applyMiddleware, compose, Store } from 'redux'
import { persistStore, persistReducer } from 'redux-persist'
import storage from 'redux-persist/lib/storage'
import { createBrowserHistory } from 'history'
import { connectRouter, routerMiddleware } from 'connected-react-router'
import thunk from 'redux-thunk'

import { rootReducer } from 'reducers'
import { ApplicationState } from 'types/schema'

// increment this number to force a refresh on the client
// i.e. - when making a breaking change to our Redux structure
const REDUX_STORE_VERSION = 1

const persistConfig = {
  key: 'root',
  storage,
  version: REDUX_STORE_VERSION,
  migrate: (state: any) => {
    if (state && state._persist.version !== REDUX_STORE_VERSION) {
      return Promise.resolve({
        config: state.config,
        lifecycle: state.lifecycle,
        _persist: state._persist,
      })
    } else {
      return Promise.resolve(state)
    }
  },
}

const persistedReducer = persistReducer(persistConfig, rootReducer)

declare const window: {
  devToolsExtension: () => any
}

export const history = createBrowserHistory()

const configureStore = (): Store<ApplicationState> => {
  const created = createStore(
    connectRouter(history)(persistedReducer),
    {},
    compose(
      applyMiddleware(routerMiddleware(history)),
      applyMiddleware(thunk),
      window.devToolsExtension ? window.devToolsExtension() : (f: any) => f
    )
  )

  const hot = (module as any).hot

  if (hot) {
    // Enable Webpack hot module replacement for reducers
    hot.accept('reducers', () => {
      const newRootReducer = require('reducers').default
      created.replaceReducer(newRootReducer())
    })
  }

  return created
}

export const store = configureStore()
export const persistor = persistStore(store)
