import configureStore from 'redux-mock-store'

import { initialState, flash, SET_FLASH, CLEAR_FLASH } from 'state/flash'

const mockStore = configureStore()

describe('reducer', () => {
  it('with SET_FLASH adds modal', () => {
    const result = flash.reducer(initialState, {
      type: SET_FLASH,
      message: 'You did a thing!',
    })
    const expected = {
      message: 'You did a thing!',
      showFlash: true,
    }
    expect(result).toEqual(expected)
  })

  it('with CLEAR_FLASH updates modal', () => {
    const result = flash.reducer(
      {
        message: 'You did a thing!',
        showFlash: true,
      },
      {
        type: CLEAR_FLASH,
      }
    )
    const expected = {
      message: '',
      color: '',
      showFlash: false,
    }
    expect(result).toEqual(expected)
  })
})

describe('set', () => {
  it('when ok, dispatch SET_FLASH, set timer, then dispatch CLEAR_FLASH', async () => {
    jest.useFakeTimers()
    const store = mockStore()

    await flash.set(store.dispatch, 'flash flash flash')

    expect(store.getActions()[0]).toEqual({
      type: SET_FLASH,
      message: 'flash flash flash',
    })

    expect(setTimeout).toHaveBeenLastCalledWith(expect.any(Function), 3000)

    jest.runAllTimers()

    expect(store.getActions()[1]).toEqual({
      type: CLEAR_FLASH,
    })
  })
})
