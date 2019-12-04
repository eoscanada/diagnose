import React, { useEffect, useState } from "react";
import { RouteComponentProps, withRouter } from "react-router";
import { MainLayout } from "../components/main-layout";
import { useAppConfig } from "../hooks/dignose";
import { BlockRange } from "../types";
import { ApiService } from "../utils/api";
import { BlockHolesList } from "../components/block-holes-list";
import { Col, Row, Typography, Tag, PageHeader, Descriptions } from "antd";
import { Btn } from "../atoms/buttons";

const { Text } = Typography;

function BaseSearchIndexesPage(props: RouteComponentProps): React.ReactElement {
  const [process, setProcess] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [ranges, setRanges] = useState<BlockRange[]>([]);

  const appConfig = useAppConfig();

  useEffect(() => {
    let stream: WebSocket;
    if (process) {
      setRanges([]);
      stream = ApiService.stream({
        route: "search_holes",
        onComplete: () => {
          setProcess(false);
        },
        onData: resp => {
          switch (resp.type) {
            case "BlockRange":
              setRanges(currentRanges => [...currentRanges, resp.payload]);
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
        title="Search Indexes"
        subTitle="hole checker"
        extra={[
          <Btn
            key={1}
            stopText="Stop Hole Checker"
            startText="Check Search Indexes Holes"
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
          <Descriptions.Item label="Index Store URL">
            {appConfig && (
              <Text code>
                <a
                  target="_blank"
                  rel="noopener noreferrer"
                  href={appConfig.indexesStoreUrl}
                >
                  {appConfig.blockStoreUrl}
                </a>
              </Text>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Shard size">
            {appConfig && (
              <Tag color="#2db7f5">shard size: {appConfig.shardSize}</Tag>
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

export const SearchIndexesPage = withRouter(BaseSearchIndexesPage);
