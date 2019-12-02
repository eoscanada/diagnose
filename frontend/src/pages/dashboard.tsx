import React, { useEffect } from "react"
import { withRouter } from "react-router"
import { MainLayout } from "../components/main-layout"
import { useAppConfig } from "../hooks/dignose"
import {ApiService} from "../utils/api";
import {PeerEvent} from "../types";

function BaseDashboard(): React.ReactElement {

  const appConfig = useAppConfig()

  return (
    <MainLayout config={appConfig}>
      <h1>Dashboard</h1>
    </MainLayout>
  )
}

export const DashboardPage = withRouter(BaseDashboard)