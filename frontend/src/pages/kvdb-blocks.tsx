import React, { useState, useEffect } from "react"
import { BlockRange } from "../types"
import { ApiService } from "../utils/api"
import { BlockHolesList } from "../components/block-holes-list"
import { MainLayout } from "../components/main-layout"
import { Typography, Row, Col, PageHeader, Descriptions, Icon } from "antd"
import { Btn } from "../atoms/buttons"
import { useStore } from "../store"
import { CreateInputModalForm } from "../components/input-modal"
import queryString from "query-string"

const { Text } = Typography

type ProcessState = "block_holes" | "validating_blocks"

const EditKvdbConnectionModal = CreateInputModalForm({ name: "kvdb_blocks_edit" })

export const KvdbBlocksPage: React.FC = () => {
  const [process, setProcess] = useState<ProcessState | undefined>()
  const [elapsed, setElapsed] = useState(0)
  const [title, setTitle] = useState("")
  const [ranges, setRanges] = useState<BlockRange[]>([])
  const [editKvdbConnectionModalVisible, setEditKvdbConnectionModalVisible] = useState(false)

  const [{ config: appConfig }, { setConfig }] = useStore()

  useEffect(() => {
    let stream: WebSocket
    if (process !== undefined) {
      setRanges([])
      setElapsed(0)

      if (process === "block_holes") {
        setTitle("Processing Block Holes")
        stream = ApiService.stream({
          route:
            "kvdb_blk_holes?" +
            queryString.stringify({ connection_info: appConfig.kvdbConnectionInfo }),

          onComplete: () => {
            setProcess(undefined)
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
      } else if (process === "validating_blocks") {
        setTitle("Validating Blocks")
        stream = ApiService.stream({
          route: "kvdb_blk_validation",
          onComplete: () => {
            setProcess(undefined)
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
    }

    return () => {
      if (stream) {
        stream.close()
      }
    }
  }, [process, appConfig.kvdbConnectionInfo])

  return (
    <MainLayout>
      <PageHeader
        title="KVDB Blocks"
        subTitle="hole checker & validator for KVDB blocks"
        extra={[
          <Btn
            key={1}
            stopText="Stop Hole Checker"
            startText="Check Block Holes"
            loading={process === "block_holes"}
            onStart={() => setProcess("block_holes")}
            onStop={() => setProcess(undefined)}
          />,
          <Btn
            key={2}
            stopText="Stop Validation"
            startText="Validate Blocks"
            loading={process === "validating_blocks"}
            onStart={(e) => {
              e.preventDefault()
              setProcess("validating_blocks")
            }}
            onStop={(e) => {
              e.preventDefault()
              setProcess(undefined)
            }}
          />
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Connection Info">
            {appConfig.kvdbConnectionInfo && <Text code>{appConfig.kvdbConnectionInfo}</Text>}
            &nbsp;
            <Icon
              type="edit"
              theme="outlined"
              onClick={() => {
                setEditKvdbConnectionModalVisible(true)
              }}
            />
            <EditKvdbConnectionModal
              initialInput={appConfig.kvdbConnectionInfo}
              visible={editKvdbConnectionModalVisible}
              onInput={(input) => {
                setConfig({ kvdbConnectionInfo: input })
                setEditKvdbConnectionModalVisible(false)
              }}
              onCancel={() => {
                setEditKvdbConnectionModalVisible(false)
              }}
            />
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>
      <Row>
        <Col>
          <h1>{title}</h1>
          <BlockHolesList ranges={ranges} inv={true} elapsed={elapsed} />
        </Col>
      </Row>
    </MainLayout>
  )
}
