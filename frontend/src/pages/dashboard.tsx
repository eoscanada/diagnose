import React from "react"
import { withRouter } from "react-router"
import { MainLayout } from "../components/main-layout"
import { useAppConfig } from "../hooks/dignose"

function BaseDashboard(): React.ReactElement {

  const appConfig = useAppConfig()

  return (
    <MainLayout config={appConfig}>
      <h1>Dashboard</h1>
    </MainLayout>
  )
}

export const DashboardPage = withRouter(BaseDashboard)