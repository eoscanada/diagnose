import React, {useState, useEffect} from "react"
import { withRouter, RouteComponentProps } from "react-router"
import { BlockRangeData } from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import {Typography, Row, Col, Button, Icon, Descriptions, PageHeader} from "antd"
import {Btn} from "../atoms/buttons";
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
      <PageHeader
        ghost={true}
        title="Block Logs"
        subTitle={"hole checker"}
        extra={[
          <Btn key={1}  stopText={'Stop Hole Checker'} startText={'Check Block Holes'} loading={process} onStart={() => setProcess(true)} onStop={() => setProcess(false)} />,
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Block Store URL">
            {
              appConfig &&
              <a target={"_blank"} href={appConfig.blockStoreUrl}>
                {appConfig.blockStoreUrl}
              </a>            }
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>
      <Row>
        <Col>
          { <BlockHolesList ranges={ranges} /> }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const BlockHolesPage = withRouter(BaseBlockHolesPage)