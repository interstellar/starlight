import { Dispatch } from 'redux'
import { LOGOUT_SUCCESS } from 'state/lifecycle'

export const Starlightd = {
  post: async (dispatch: Dispatch, url = ``, data = {}) => {
    let response
    try {
      // Default options marked with *
      response = await fetch(url, {
        method: 'POST',
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, same-origin, *omit
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
        body: JSON.stringify(data), // body data type must match "Content-Type"
      })

      if (response.status === 401) {
        dispatch({ type: LOGOUT_SUCCESS })
      }

      if ((response.headers.get('content-type') || '').includes('json')) {
        return { body: await response.json(), ok: response.ok }
      } else {
        return { body: '', ok: response.ok }
      }
    } catch (error) {
      return { body: '', ok: false }
    }
  },
}
