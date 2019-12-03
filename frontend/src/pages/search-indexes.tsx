import React, {useEffect, useState} from "react"
import {RouteComponentProps, withRouter} from "react-router"
import { MainLayout } from "../components/main-layout"
import { useAppConfig } from "../hooks/dignose"
import {BlockRangeData} from "../types";
import {ApiService} from "../utils/api";
import {BlockHolesList} from "../components/block-holes-list";
import {Button, Col, Icon, Row, Typography, Tag, PageHeader, Descriptions} from "antd";
import {Btn} from "../atoms/buttons";
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
      <PageHeader
        ghost={true}
        title="Search Indexes"
        subTitle={"hole checker"}
        extra={[
          <Btn key={1}  stopText={'Stop Hole Checker'} startText={'Check Search Indexes Holes'} loading={process} onStart={() => setProcess(true)} onStop={() => setProcess(false)} />,
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Index Store URL">
            {
              appConfig &&
              <Text code>
                <a target={"_blank"} href={appConfig.indexesStoreUrl}>
                  {appConfig.blockStoreUrl}
                </a>
              </Text>         }
          </Descriptions.Item>
          <Descriptions.Item label="Shard size">
            {
              appConfig &&
              <Tag color="#2db7f5">shard size:  {appConfig.shardSize}</Tag>
            }
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

export const SearchIndexesPage = withRouter(BaseSearchIndexesPage)