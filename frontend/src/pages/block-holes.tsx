import React, { useState, useEffect } from "react";
import { withRouter, RouteComponentProps } from "react-router";
import { Row, Col, Descriptions, PageHeader } from "antd";
import { BlockRange } from "../types";
import { ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose";
import { BlockHolesList } from "../components/block-holes-list";
import { MainLayout } from "../components/main-layout";
import { Btn } from "../atoms/buttons";

function BaseBlockHolesPage(props: RouteComponentProps): React.ReactElement {
  const [process, setProcess] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [ranges, setRanges] = useState<BlockRange[]>([]);

  const appConfig = useAppConfig();

  useEffect(() => {
    let stream: WebSocket;
    if (process) {
      setRanges([]);
      setElapsed(0);
      stream = ApiService.stream({
        route: "block_holes",
        onComplete: () => {
          setProcess(false);
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

    return () => {
      if (stream) {
        stream.close();
      }
    };
  }, [process]);

  return (
    <MainLayout config={appConfig} {...props}>
      <PageHeader
        title="Block Logs"
        subTitle="hole checker"
        extra={[
          <Btn
            key={1}
            stopText="Stop Hole Checker"
            startText="Check Block Holes"
            loading={process}
            onStart={e => {
              e.preventDefault();
              setProcess(true);
            }}
            onStop={e => {
              e.preventDefault();
              setProcess(false);
            }}
          />
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Block Store URL">
            {appConfig && (
              <a
                target="_blank"
                rel="noopener noreferrer"
                href={appConfig.blockStoreUrl}
              >
                {appConfig.blockStoreUrl}
              </a>
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
  );
}

export const BlockHolesPage = withRouter(BaseBlockHolesPage);
