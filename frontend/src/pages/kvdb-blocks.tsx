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

  const VALIDATE_BLOCKS = "validating_blocks";
  const BLOCK_HOLE = "block_holes";

  const [process, setProcess] = useState("")
  const [title, setTitle] = useState("")
  const [ranges,setRanges] = useState<BlockRangeData[]>([])

  const appConfig = useAppConfig()

  const processingBlockHoles = ():boolean   => {
    return (process === BLOCK_HOLE)
  }

  const validatingBlocks = ():boolean => {
    return (process === VALIDATE_BLOCKS)
  }

  useEffect(
    () => {
      var stream:WebSocket;

      if(process !== "") {
        if (processingBlockHoles()) {
          setTitle("Processing Block Holes")
          stream = ApiService.stream<BlockRangeData>({
            route: "kvdb_blk_holes",
            onComplete: function () {
              setProcess("")
            },
            onData: (resp)  => {
              setRanges((ranges) => [...ranges, resp.data])
            }
          })
        } else if(validatingBlocks()) {
          setTitle("Validating Blocks")
          stream = ApiService.stream<BlockRangeData>({
            route: "kvdb_blk_validation",
            onComplete: function () {
              setProcess("")
            },
            onData: (resp)  => {
              setRanges((ranges) => [...ranges, resp.data])
            }
          })

        }
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
          <Button key={1} type="primary" shape="round" icon="play-circle" size={'default'} loading={processingBlockHoles()} onClick={() =>setProcess(BLOCK_HOLE)} >
            Check Block Holes
          </Button>,
          <Button key={2} type="primary" shape="round" icon="play-circle" size={'default'} loading={validatingBlocks()} onClick={() =>setProcess(VALIDATE_BLOCKS)} >
            Validate Blocks
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
          <h1>{title}</h1>
          { <BlockHolesList  ranges={ranges}  inv={true} /> }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const KvdbBlocksPage = withRouter(BaseKvdbBlocksPage)