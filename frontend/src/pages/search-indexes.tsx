import React, {useState} from "react"
import { withRouter } from "react-router"
import { MainLayout } from "../components/main-layout"
import { useAppConfig } from "../hooks/dignose"
import {BlockRangeData} from "../types";
import {ApiService} from "../utils/api";
import {BlockHolesList} from "../components/block-holes-list";
import {Button, Col, Icon, Row, Typography, Tag} from "antd";
const { Text } = Typography;

function BaseSearchIndexesPage(): React.ReactElement {

  const [processing, setProcessing] = useState(false)
  const [ranges,setRanges] = useState<BlockRangeData[]>([])

  const appConfig = useAppConfig()

  const loadIndexes = () => {
    setProcessing(true)
    setRanges([])
    ApiService.stream<BlockRangeData>({
      route: "api/search_holes",
      onComplete: function () {
        setProcessing(false)
      },
      onData: (resp)  => {
        setRanges((ranges) => [...ranges, resp.data])
      }
    })
  };

  return (
    <MainLayout config={appConfig}>
      <Row justify="space-between">
        <Col span={12} style={{ textAlign: "left"}}>
          <h1>Checking holes in Search indexes</h1>
        </Col>
        <Col span={12} style={{ textAlign: "right"}}>
          <Button type="primary" loading={processing} onClick={loadIndexes}>
            process indexes
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
                Index Store URL:
                <Text code>
                  <a target={"_blank"} href={appConfig.indexesStoreUrl}>
                    {appConfig.blockStoreUrl}
                  </a>
                </Text>
                <Tag color="#2db7f5">shard size:  {appConfig.shardSize}</Tag>


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

export const SearchIndexesPage = withRouter(BaseSearchIndexesPage)