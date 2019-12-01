import React from "react"
import { LogoDfuse } from "../atoms/svg"
import { DiagnoseConfig } from '../types'
import { Row, Col, Tag } from 'antd';


export function TopHeader(props: {
  config: DiagnoseConfig | undefined,
}): React.ReactElement {

  let protocol = ""
  let namespace = ""
  if (props.config) {
    protocol = props.config.protocol
    namespace = props.config.namespace
  }

  return (
    <div>
      <Row justify="space-between">
        <Col span={8}>
          <LogoDfuse fill="#EC5664" height="44px" />
        </Col>
        <Col span={8}  style={{ textAlign: "center"}}>
          Diagnose { protocol }
        </Col>
        <Col span={8}  style={{ textAlign: "right"}}>
          <Tag>{namespace}</Tag>
        </Col>
      </Row>
    </div>
  )
}