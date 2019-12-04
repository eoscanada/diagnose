import React, { useState, useEffect } from "react";
import { withRouter, RouteComponentProps } from "react-router";
import { BlockRange } from "../types";
import { ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose";
import { BlockHolesList } from "../components/block-holes-list";
import { MainLayout } from "../components/main-layout";
import { Typography, Row, Col, PageHeader, Descriptions } from "antd";
import { Btn } from "../atoms/buttons";

const { Text } = Typography;

function BaseKvdbBlocksPage(props: RouteComponentProps): React.ReactElement {
  const VALIDATE_BLOCKS = "validating_blocks";
  const BLOCK_HOLE = "block_holes";

  const [process, setProcess] = useState("");
  const [elapsed, setElapsed] = useState(0);
  const [title, setTitle] = useState("");
  const [ranges, setRanges] = useState<BlockRange[]>([]);

  const appConfig = useAppConfig();

  const processingBlockHoles = (): boolean => {
    return process === BLOCK_HOLE;
  };

  const validatingBlocks = (): boolean => {
    return process === VALIDATE_BLOCKS;
  };

  useEffect(() => {
    let stream: WebSocket;

    if (process !== "") {
      setRanges([]);
      setElapsed(0);
      if (processingBlockHoles()) {
        setTitle("Processing Block Holes");
        stream = ApiService.stream({
          route: "kvdb_blk_holes",
          onComplete: () => {
            setProcess("");
          },
          onData: resp => {
            switch (resp.type) {
              case "Transaction":
                break;
              case "BlockRange":
                setRanges(currentRanges => [...currentRanges, resp.payload]);
                break;
              case "Message":
                break;
              case "Progress":
                setElapsed(resp.payload.elapsed);
                break;
            }
          }
        });
      } else if (validatingBlocks()) {
        setTitle("Validating Blocks");
        stream = ApiService.stream({
          route: "kvdb_blk_validation",
          onComplete: () => {
            setProcess("");
          },
          onData: resp => {
            switch (resp.type) {
              case "Transaction":
                break;
              case "BlockRange":
                setRanges(currentRanges => [...currentRanges, resp.payload]);
                break;
              case "Message":
                break;
              case "Progress":
                setElapsed(resp.payload.elapsed);
                break;
            }
          }
        });
      }
    }

    return () => {
      if (stream) {
        stream.close();
      }
    };
  }, [process]);

  return (
    <MainLayout config={appConfig} {...props}>
      <PageHeader
        title="KVDB Blocks"
        subTitle="hole checker & validator for KVDB blocks"
        extra={[
          <Btn
            key={1}
            stopText="Stop Hole Checker"
            startText="Check Block Holes"
            loading={processingBlockHoles()}
            onStart={() => setProcess(BLOCK_HOLE)}
            onStop={() => setProcess("")}
          />,
          <Btn
            key={2}
            stopText="Stop Validation"
            startText="Validate Blocks"
            loading={validatingBlocks()}
            onStart={e => {
              e.preventDefault();
              setProcess(VALIDATE_BLOCKS);
            }}
            onStop={e => {
              e.preventDefault();
              setProcess("");
            }}
          />
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Connection Info">
            {appConfig && <Text code>{appConfig.kvdbConnectionInfo}</Text>}
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
  );
}

export const KvdbBlocksPage = withRouter(BaseKvdbBlocksPage);
