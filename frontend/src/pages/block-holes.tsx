import React, { useState, useEffect } from "react"
import { Row, Col, Descriptions, PageHeader, Icon } from "antd"
import { BlockRange } from "../types"
import { ApiService } from "../utils/api"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import { Btn } from "../atoms/buttons"
import { CreateInputModalForm } from "../components/input-modal"
import queryString from "query-string"
import { useStore } from "../store"

const EditBlocksUrlModal = CreateInputModalForm({ name: "blocks_url_edit " })

export const BlockHolesPage: React.FC = () => {
  const [process, setProcess] = useState(false)
  const [elapsed, setElapsed] = useState(0)
  const [ranges, setRanges] = useState<BlockRange[]>([])
  const [editUrlModalVisible, setEditUrlModalVisible] = useState(false)

  const [{ config: appConfig }, { setConfig }] = useStore()

  useEffect(() => {
    let stream: WebSocket
    if (process) {
      setRanges([])
      setElapsed(0)
      stream = ApiService.stream({
        route: "block_holes?" + queryString.stringify({ blocks_url: appConfig.blockStoreUrl }),
        onComplete: () => {
          setProcess(false)
        },
        onData: (resp) => {
          switch (resp.type) {
            case "Transaction":
              break
            case "BlockRange":
              setRanges((currentRanges) => [...currentRanges, resp.payload])
              break
            case "Message":
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
  }, [process, appConfig.blockStoreUrl])

  return (
    <MainLayout>
      <PageHeader
        title="Block Logs"
        subTitle="hole checker"
        extra={[
          <Btn
            key={1}
            stopText="Stop Hole Checker"
            startText="Check Block Holes"
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
          <Descriptions.Item label="Block Store URL">
            {appConfig.blockStoreUrl && (
              <a target="_blank" rel="noopener noreferrer" href={appConfig.blockStoreUrl}>
                {appConfig.blockStoreUrl}
              </a>
            )}
            &nbsp;
            <Icon
              type="edit"
              theme="outlined"
              onClick={() => {
                setEditUrlModalVisible(true)
              }}
            />
            <EditBlocksUrlModal
              initialInput={appConfig.blockStoreUrl}
              visible={editUrlModalVisible}
              onInput={(input) => {
                setConfig({ blockStoreUrl: input })
                setEditUrlModalVisible(false)
              }}
              onCancel={() => {
                setEditUrlModalVisible(false)
              }}
            />
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
