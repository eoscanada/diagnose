import React, { createContext, useContext, useEffect, useState } from "react"
import { DiagnoseConfig } from "./types"
import { diagnoseService } from "./services/diganose-config"

type GlobalState = {
  config: DiagnoseConfig
  error?: any
}

interface Controller {
  setConfig(config: DiagnoseConfig): void
  setError(error: any): void
}

const initialState: GlobalState = { config: {} }

const controllerFactory = (
  setState: React.Dispatch<React.SetStateAction<GlobalState>>
): Controller => {
  return {
    setConfig(config: DiagnoseConfig) {
      setState((prev: GlobalState) => ({ ...prev, config: { ...prev.config, ...config } }))
    },

    setError(error: any) {
      setState((prev: GlobalState) => ({ ...prev, error }))
    }
  }
}

const Context = createContext<[GlobalState, Controller]>([
  initialState,
  controllerFactory(() => {})
])

export const Store: React.FC = ({ children }) => {
  const [state, setState] = useState<GlobalState>(initialState)
  const controller = controllerFactory(setState)

  useEffect(
    () => {
      diagnoseService.config().then((response) => {
        console.log("Updating config from received from server", response)
        if (response.type === "data") {
          controller.setConfig(response.data)
        } else {
          console.log("Unable to fetch config from server", response.errors)
          controller.setError(response.errors)
        }
      })
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

  return <Context.Provider value={[state, controller]}>{children}</Context.Provider>
}

export const useStore = () => useContext(Context)
