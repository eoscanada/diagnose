import React, {useEffect, useState} from "react"
import {RouteComponentProps, withRouter} from "react-router"
import { MainLayout } from "../components/main-layout"
import { useAppConfig } from "../hooks/dignose"
import {BlockRangeData} from "../types";
import {ApiService} from "../utils/api";
import {BlockHolesList} from "../components/block-holes-list";
import {Button, Col, Icon, Row, Typography, Tag} from "antd";
const { Text } = Typography;

function BaseSearchIndexesPage(
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
          route: "search_holes",
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
          <h1>Search Indexes hole checker</h1>
        </Col>
        <Col span={12} style={{ textAlign: "right"}}>
          <Button type="primary" loading={process} onClick={() =>setProcess(!process)}>
            process indexe
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