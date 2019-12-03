import React, {useState, useEffect} from "react"
import { withRouter, RouteComponentProps } from "react-router"
import { BlockRangeData } from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import {Typography, Row, Col, Button, Icon} from "antd"
const { Text } = Typography;

function BaseKvdbTrxsPage(
  props: RouteComponentProps
): React.ReactElement {

  const [process, setProcess] = useState(false)
  const appConfig = useAppConfig()



  return (
    <MainLayout config={appConfig} {...props}>
      <Row justify="space-between">
        <Col span={12} style={{ textAlign: "left"}}>
          <h1>KVDB Transaction Checker</h1>
        </Col>
        <Col span={12} style={{ textAlign: "right"}}>
          <Button type="primary" loading={process} onClick={() =>setProcess(!process)}>
            process trxs
            <Icon type="monitor" />
          </Button>
        </Col>
      </Row>
      <Row>
        <Col>
          {
            appConfig &&
            (
              <div>
              </div>
            )
          }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const KvdbTrxsPage = withRouter(BaseKvdbTrxsPage)