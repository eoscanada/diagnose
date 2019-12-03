import React, { useEffect, useState } from "react";
import { withRouter, RouteComponentProps } from "react-router";
import { Row, Col, Typography, PageHeader, Descriptions } from "antd";
import { Peer } from "../types";
import { ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose";
import { SearchPeerList } from "../components/search-peer-list";
import { MainLayout } from "../components/main-layout";
import { Btn } from "../atoms/buttons";

const { Text } = Typography;

function BaseDmeshPage(props: RouteComponentProps): React.ReactElement {
  const [track, setTrack] = useState(false);
  const [peers, setPeers] = useState<Peer[]>([]);

  const appConfig = useAppConfig();

  const peerIndex = (peerHost: string): number => {
    return peers.findIndex(peer => peer.host === peerHost);
  };

  const deletePeer = (peer: Peer) => {
    const index = peerIndex(peer.host);
    if (index >= 0) {
      setPeers(currentPeers => {
        return [...currentPeers.slice(index, 1)];
      });
    }
  };

  const updatePeer = (peer: Peer) => {
    setPeers(currentPeers => {
      const newCUrrentPeers = currentPeers.map(peerItem => {
        if (peerItem.host === peer.host) {
          return peer;
        }
        return peerItem;
      });
      return newCUrrentPeers;
    });
  };

  const addPeer = (peer: Peer) => {
    setPeers(currentPeers => {
      return [...currentPeers, peer];
    });
  };

  useEffect(() => {
    let stream: WebSocket;
    if (track) {
      setPeers([]);
      stream = ApiService.stream({
        route: "search_peers",
        onComplete: () => {
          setTrack(false);
        },
        onData: resp => {
          switch (resp.type) {
            case "Transaction":
              break;
            case "BlockRange":
              break;
            case "Message":
              break;
            case "PeerEvent":
              if (resp.payload.EventName === "sync") {
                addPeer(resp.payload.Peer);
              } else if (resp.payload.EventName === "update") {
                updatePeer(resp.payload.Peer);
              } else if (resp.payload.EventName === "delete") {
                deletePeer(resp.payload.Peer);
              }
              break;
          }
        }
      });
    }

    return () => {
      if (stream) {
        console.log("closing stream");
        stream.close();
      }
    };
  }, [track]);

  return (
    <MainLayout config={appConfig} {...props}>
      <PageHeader
        title="Dmesh Peers"
        extra={[
          <Btn
            key={1}
            stopText="Stop Tracking Dmesh"
            startText="Track Dmesh"
            loading={track}
            onStart={() => setTrack(true)}
            onStop={() => setTrack(false)}
          />
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Watch Key">
            {appConfig && (
              <Text code>
                /{appConfig.namespace}/{appConfig.dmeshServiceVersion}/search
              </Text>
            )}
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>
      <Row>
        <Col>
          <div style={{ marginTop: "10px" }}>
            <SearchPeerList peers={peers} />
          </div>
        </Col>
      </Row>
    </MainLayout>
  );
}

export const DmeshPage = withRouter(BaseDmeshPage);
