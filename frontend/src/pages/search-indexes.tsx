import React, { useEffect, useState } from "react";
import { RouteComponentProps, withRouter } from "react-router";

import { MainLayout } from "../components/main-layout";
import { useAppConfig } from "../hooks/dignose";
import { BlockRange } from "../types";
import { ApiService } from "../utils/api";
import { BlockHolesList } from "../components/block-holes-list";
import { Col, Row, Typography, PageHeader, Descriptions, Select } from "antd";
import { Btn } from "../atoms/buttons";

const { Option } = Select;
const { Text } = Typography;

function BaseSearchIndexesPage(props: RouteComponentProps): React.ReactElement {
  const [process, setProcess] = useState(false);
  const [shardSize, setShardSize] = useState(5000);
  const [elapsed, setElapsed] = useState(0);
  const [ranges, setRanges] = useState<BlockRange[]>([]);

  const appConfig = useAppConfig();

  useEffect(() => {
    let stream: WebSocket;
    if (process) {
      setRanges([]);
      stream = ApiService.stream({
        route: `search_holes?shard_size=${shardSize}`,
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
  }, [process, shardSize]);

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
              <>
                <Select
                  defaultValue={shardSize}
                  style={{ width: 120 }}
                  onChange={(value: number) => {
                    setShardSize(value);
                  }}
                >
                  {appConfig.shardSizes.map(ss => {
                    return <Option value={ss}>{ss}</Option>;
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
  );
}

export const SearchIndexesPage = withRouter(BaseSearchIndexesPage);
