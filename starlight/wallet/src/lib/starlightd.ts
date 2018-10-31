import { post } from 'client/client'
import { LOGOUT_SUCCESS } from 'state/lifecycle'

export const Starlightd = {
  post: async (dispatch: any, url = ``, data = {}) => {
    const response = await post(url, data)
    if (!response.loggedIn) {
      dispatch({
        type: LOGOUT_SUCCESS,
      })
    }
    return response
  },
}
