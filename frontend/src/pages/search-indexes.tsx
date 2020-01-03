import React, { useEffect, useState } from "react"
import { MainLayout } from "../components/main-layout"
import { BlockRange } from "../types"
import { ApiService } from "../utils/api"
import { BlockHolesList } from "../components/block-holes-list"
import { Col, Row, Typography, PageHeader, Descriptions, Select, Icon } from "antd"
import { Btn } from "../atoms/buttons"
import { useStore } from "../store"
import queryString from "query-string"
import { CreateInputModalForm } from "../components/input-modal"

const { Option } = Select
const { Text } = Typography

const EditIndexesUrlModal = CreateInputModalForm({ name: "search_url_edit " })

export const SearchIndexesPage: React.FC = () => {
  const [process, setProcess] = useState(false)
  const [shardSize, setShardSize] = useState(5000)
  const [elapsed, setElapsed] = useState(0)
  const [ranges, setRanges] = useState<BlockRange[]>([])
  const [editUrlModalVisible, setEditUrlModalVisible] = useState(false)

  const [{ config: appConfig }, { setConfig }] = useStore()

  useEffect(() => {
    let stream: WebSocket
    if (process) {
      setRanges([])
      stream = ApiService.stream({
        route:
          "search_holes?" +
          queryString.stringify({ indexes_url: appConfig.indexesStoreUrl, shard_size: shardSize }),
        onComplete: () => {
          setProcess(false)
        },
        onData: (resp) => {
          switch (resp.type) {
            case "BlockRange":
              setRanges((currentRanges) => [...currentRanges, resp.payload])
              break
            case "Progress":
              setElapsed(resp.payload.elapsed)
              break
          }
        }
      })
    }

    return () => {
      if (stream) {
        stream.close()
      }
    }
  }, [process, shardSize, appConfig.indexesStoreUrl])

  return (
    <MainLayout>
      <PageHeader
        title="Search Indexes"
        subTitle="hole checker"
        extra={[
          <Btn
            key={1}
            stopText="Stop Hole Checker"
            startText="Check Search Indexes Holes"
            loading={process}
            onStart={(e) => {
              e.preventDefault()
              setProcess(true)
            }}
            onStop={(e) => {
              e.preventDefault()
              setProcess(false)
            }}
          />
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Index Store URL">
            {appConfig.indexesStoreUrl && (
              <Text code>
                <a target="_blank" rel="noopener noreferrer" href={appConfig.indexesStoreUrl}>
                  {appConfig.indexesStoreUrl}
                </a>
              </Text>
            )}
            &nbsp;
            <Icon
              type="edit"
              theme="outlined"
              onClick={() => {
                setEditUrlModalVisible(true)
              }}
            />
            <EditIndexesUrlModal
              initialInput={appConfig.indexesStoreUrl}
              visible={editUrlModalVisible}
              onInput={(input) => {
                setConfig({ indexesStoreUrl: input })
                setEditUrlModalVisible(false)
              }}
              onCancel={() => {
                setEditUrlModalVisible(false)
              }}
            />
          </Descriptions.Item>
          <Descriptions.Item label="Shard size">
            {appConfig.shardSizes && (
              <>
                <Select
                  defaultValue={shardSize}
                  style={{ width: 120 }}
                  onChange={(value: number) => {
                    setShardSize(value)
                  }}
                >
                  {appConfig.shardSizes.map((ss) => {
                    return <Option value={ss}>{ss}</Option>
                  })}
                </Select>
              </>
            )}
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>
      <Row>
        <Col>
          <BlockHolesList ranges={ranges} elapsed={elapsed} />
        </Col>
      </Row>
    </MainLayout>
  )
}
