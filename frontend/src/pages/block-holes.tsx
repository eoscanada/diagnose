import React, { useState } from "react"
import { withRouter, RouteComponentProps } from "react-router"
import { BlockRangeData } from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import {Typography, Row, Col, Button, Icon, List} from "antd"
const { Text } = Typography;

function BaseBlockHolesPage(
  props: RouteComponentProps
): React.ReactElement {

  const [processing, setProcessing] = useState(false)
  const [ranges,setRanges] = useState<BlockRangeData[]>([])

  const appConfig = useAppConfig()

  const loadBlocks = () => {
    setProcessing(true)
    setRanges([])
    ApiService.stream<BlockRangeData>({
      route: "block_holes",
      onComplete: function () {
        setProcessing(false)
      },
      onData: (resp)  => {
        console.log("new data point: ", resp.data)
        setRanges((ranges) => [...ranges, resp.data])
      }
    })
  };

  return (
    <MainLayout config={appConfig}>
      <Row justify="space-between">
        <Col span={12} style={{ textAlign: "left"}}>
          <h1>Block Logs Hole Checker</h1>
        </Col>
        <Col span={12} style={{ textAlign: "right"}}>
          <Button type="primary" loading={processing} onClick={loadBlocks}>
            process block
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
                <p>
                  Block Store URL:
                  <Text code>
                    <a target={"_blank"} href={appConfig.blockStoreUrl}>
                      {appConfig.blockStoreUrl}
                    </a>
                 </Text>
                </p>
                <BlockHolesList ranges={ranges} />
              </div>
            )
          }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const BlockHolesPage = withRouter(BaseBlockHolesPage)