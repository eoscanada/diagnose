import React, {useState, useEffect} from "react"
import { withRouter, RouteComponentProps } from "react-router"
import { BlockRangeData } from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import {Typography, Row, Col, Button, Icon} from "antd"
const { Text } = Typography;

function BaseBlockHolesPage(
  props: RouteComponentProps
): React.ReactElement {

  const [process, setProcess] = useState(false)
  const [ranges,setRanges] = useState<BlockRangeData[]>([])

  const appConfig = useAppConfig()


  useEffect(
    () => {
      var stream:WebSocket;
      if (process) {
        setRanges([])
        stream = ApiService.stream<BlockRangeData>({
          route: "block_holes",
          onComplete: function () {
            setProcess(false)
          },
          onData: (resp)  => {
            setRanges((ranges) => [...ranges, resp.data])
          }
        })
      }

      return () => {
        if(stream) {
          stream.close()
        }
      }
    },
    [process],
  );


  return (
    <MainLayout config={appConfig} {...props}>
      <Row justify="space-between">
        <Col span={12} style={{ textAlign: "left"}}>
          <h1>Block Logs Hole Checker</h1>
        </Col>
        <Col span={12} style={{ textAlign: "right"}}>
          <Button type="primary" loading={process} onClick={() =>setProcess(!process)}>
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