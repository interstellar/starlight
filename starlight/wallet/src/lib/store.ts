import { createStore, applyMiddleware, compose, Store } from 'redux'
import { persistStore, persistReducer } from 'redux-persist'
import storage from 'redux-persist/lib/storage'
import { createBrowserHistory } from 'history'
import { connectRouter, routerMiddleware } from 'connected-react-router'
import thunk from 'redux-thunk'

import { rootReducer } from 'reducers'
import { ApplicationState } from 'types/schema'

const persistConfig = {
  key: 'root',
  storage,
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
