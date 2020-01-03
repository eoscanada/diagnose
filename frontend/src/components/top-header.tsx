import React from "react"
import { Row, Col, Tag } from "antd"
import { LogoDfuse } from "../atoms/svg"
import { useStore } from "../store"

export const TopHeader: React.FC = () => {
  const [store] = useStore()

  return (
    <div>
      <Row type="flex" justify="space-between" align="middle">
        <Col span={8}>
          <LogoDfuse fill="#EC5664" height="44px" />
        </Col>
        <Col span={8} style={{ textAlign: "center" }}>
          Diagnose {store.config.protocol || "Unknwon"}
        </Col>
        <Col span={8} style={{ textAlign: "right" }}>
          <Tag>{store.config.namespace || "Unknown"}</Tag>
        </Col>
      </Row>
    </div>
  )
}
