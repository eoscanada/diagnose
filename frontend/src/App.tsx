import React from "react"
import { Router } from "./router"
import { Store } from "./store"

function App() {
  return (
    <Store>
      <Router />
    </Store>
  )
}

export default App
