import React, { useState, useEffect } from "react";
import { withRouter, RouteComponentProps } from "react-router";
import { SocketMessage } from "../types";
import { ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose";
import { MainLayout } from "../components/main-layout";
import {
  Typography,
  Row,
  Col,
  Icon,
  PageHeader,
  Descriptions,
  List
} from "antd";
import { Btn } from "../atoms/buttons";
import { formatNanoseconds } from "../utils/format";

const { Text } = Typography;

function BaseKvdbTrxsPage(props: RouteComponentProps): React.ReactElement {
  const [process, setProcess] = useState(false);
  const [items, setItems] = useState<SocketMessage[]>([]);
  const [elapsed, setElapsed] = useState(0);
  const appConfig = useAppConfig();

  useEffect(() => {
    let stream: WebSocket;

    if (process) {
      setItems([]);
      setElapsed(0);
      stream = ApiService.stream({
        route: "kvdb_trx_validation",
        onComplete: () => {
          console.log("completed kvdb_trx_validation");
          setProcess(false);
        },
        onData: resp => {
          switch (resp.type) {
            case "Transaction":
              setItems(currentItems => [...currentItems, resp]);
              break;
            case "Message":
              setItems(currentItems => [...currentItems, resp]);
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

  let header = <div />;
  if (items.length > 0) {
    header = <div>Receiving Transactions...</div>;
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
            Transaction{" "}
            <a href={`https://eosq.app/${item.payload.id}`}>
              {item.payload.prefix}
            </a>{" "}
            @ #{item.payload.blockNum} missing meta:written column
          </List.Item>
        );
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
        );
    }
  };

  return (
    <MainLayout config={appConfig} {...props}>
      <PageHeader
        title="KVDB Transaction"
        subTitle="validator for KVDB transaction"
        extra={[
          <Btn
            key={1}
            stopText="Stop Trx Validation"
            startText="Check Transaction Validation"
            loading={process}
            onStart={event => {
              setProcess(true);
              event.preventDefault();
            }}
            onStop={event => {
              setProcess(false);
              event.preventDefault();
            }}
          />
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Connection Info">
            {appConfig && <Text code>{appConfig.kvdbConnectionInfo}</Text>}
          </Descriptions.Item>
          <Descriptions.Item label="Elapsed Time">
            {appConfig && (
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
            renderItem={item => renderItem(item)}
          />
        </Col>
      </Row>
    </MainLayout>
  );
}

export const KvdbTrxsPage = withRouter(BaseKvdbTrxsPage);
