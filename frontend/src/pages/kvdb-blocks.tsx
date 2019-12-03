import React, {useState, useEffect} from "react"
import { withRouter, RouteComponentProps } from "react-router"
import { BlockRangeData } from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import {Typography, Row, Col, Button, Icon, PageHeader, Descriptions } from "antd"
import { Btn } from "../atoms/buttons";
import { IconTricorder } from "../atoms/svg";

const { Text } = Typography;

function BaseKvdbBlocksPage(
  props: RouteComponentProps
): React.ReactElement {

  const [process, setProcess] = useState(false)
  const [ranges,setRanges] = useState<BlockRangeData[]>([])

  const appConfig = useAppConfig()

  useEffect(
    () => {
      var stream:WebSocket;
      if(process) {
        stream = ApiService.stream<BlockRangeData>({
          route: "kvdb_blk_holes",
          onComplete: function () {
            setProcess(false)
          },
          onData: (resp)  => {
            console.log(resp)
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
    [process]
  );

  return (
    <MainLayout config={appConfig} {...props}>
      <PageHeader
        ghost={true}
        title="KVDB Blocks"
        subTitle={"hole checker & validator for KVDB blocks"}
        extra={[
          <Btn key={1} type="primary" loading={process} onClick={() =>setProcess(!process)} icon={<IconTricorder />}>
            process blocks Holes
          </Btn>,
          <Button key={2} type="primary" loading={process} onClick={() =>setProcess(!process)}>
          process blocks validation <Icon type="monitor" />
          </Button>,
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Connection Info">
            {
              appConfig &&
              <Text code>{appConfig.kvdbConnectionInfo}</Text>
            }
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>

      <Row>
        <Col>
          <Button type="primary" shape="round" icon="download" size={'default'}>
            Download
          </Button>

          {
            <BlockHolesList
              ranges={ranges}
              inv={true}
            />
          }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const KvdbBlocksPage = withRouter(BaseKvdbBlocksPage)