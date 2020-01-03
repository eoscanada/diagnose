import React, { useState, useEffect } from "react"
import { SocketMessage } from "../types"
import { ApiService } from "../utils/api"
import { MainLayout } from "../components/main-layout"
import { Typography, Row, Col, Icon, PageHeader, Descriptions, List } from "antd"
import { Btn } from "../atoms/buttons"
import { formatNanoseconds } from "../utils/format"
import { useStore } from "../store"
import { CreateInputModalForm } from "../components/input-modal"
import queryString from "query-string"

const { Text } = Typography

const EditKvdbConnectionModal = CreateInputModalForm({ name: "kvdb_trxs_edit" })

export const KvdbTrxsPage: React.FC = () => {
  const [process, setProcess] = useState(false)
  const [items, setItems] = useState<SocketMessage[]>([])
  const [elapsed, setElapsed] = useState(0)
  const [editKvdbConnectionModalVisible, setEditKvdbConnectionModalVisible] = useState(false)

  const [{ config: appConfig }, { setConfig }] = useStore()

  useEffect(() => {
    let stream: WebSocket

    if (process) {
      setItems([])
      setElapsed(0)
      stream = ApiService.stream({
        route:
          "kvdb_trx_validation?" +
          queryString.stringify({ connection_info: appConfig.kvdbConnectionInfo }),

        onComplete: () => {
          setProcess(false)
        },
        onData: (resp) => {
          switch (resp.type) {
            case "Transaction":
              setItems((currentItems) => [...currentItems, resp])
              break
            case "Message":
              setItems((currentItems) => [...currentItems, resp])
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
  }, [process, appConfig.kvdbConnectionInfo])

  let header = <div />
  if (items.length > 0) {
    header = <div>Receiving Transactions...</div>
  }

  const renderItem = (item: SocketMessage) => {
    switch (item.type) {
      case "Transaction":
        return (
          <List.Item>
            <Icon
              style={{ fontSize: "24px" }}
              type="close-circle"
              theme="twoTone"
              twoToneColor="#f5222d"
            />
            Transaction <a href={`https://eosq.app/${item.payload.id}`}>{item.payload.prefix}</a> @
            #{item.payload.blockNum} missing meta:written column
          </List.Item>
        )
      case "Message":
        return (
          <List.Item>
            <p
              style={{
                overflowWrap: "break-word",
                flexWrap: "wrap",
                overflow: "auto"
              }}
            >
              {item.payload.message}
            </p>
          </List.Item>
        )
    }
  }

  return (
    <MainLayout>
      <PageHeader
        title="KVDB Transaction"
        subTitle="validator for KVDB transaction"
        extra={[
          <Btn
            key={1}
            stopText="Stop Trx Validation"
            startText="Check Transaction Validation"
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
          <Descriptions.Item label="Elapsed Time">
            {appConfig.kvdbConnectionInfo && (
              <span
                style={{
                  float: "right"
                }}
              >
                elapsed: {formatNanoseconds(elapsed || 0)}
              </span>
            )}
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>
      <Row>
        <Col>
          <List
            size="small"
            header={header}
            bordered
            dataSource={items}
            renderItem={(item) => renderItem(item)}
          />
        </Col>
      </Row>
    </MainLayout>
  )
}
