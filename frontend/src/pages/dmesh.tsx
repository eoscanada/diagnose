import React, {useEffect, useState} from "react"
import { withRouter, RouteComponentProps } from "react-router"
import {PeerEvent, Peer, BlockRangeData} from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { SearchPeerList } from "../components/search-peer-list"
import { MainLayout } from "../components/main-layout"
import { IconTricorder } from "../atoms/svg"
import {Row, Col, Button, Icon, Typography, PageHeader, Descriptions} from "antd"
import {Btn} from "../atoms/buttons";
const { Text } = Typography;

function BaseDmeshPage(
  props: RouteComponentProps
): React.ReactElement {

  const [track, setTrack] = useState(false)
  const [peers, setPeers] = useState<Peer[]>([])

  const appConfig = useAppConfig()

  const peerIndex = (peerHost :string):number => {
    var count = 0
    peers.map((peer) => {
      if (peer.host === peerHost) {
        return count
      } else {
        count ++
      }
    });
    return -1
  }

  const deletePeer = (peer: Peer) => {
    var index = peerIndex(peer.host)
    if (index >= 0) {
      setPeers( (currentPeers) => {
        return [...currentPeers.slice(index,1)]
      })
    }
  }

  const updatePeer = (peer: Peer) => {
    setPeers((currentPeers) => {
      const newCUrrentPeers = currentPeers.map((peerItem) => {
        if (peerItem.host === peer.host) {
          return peer
        } else {
          return peerItem;
        }
      });
      return newCUrrentPeers
    });
  }

  const addPeer = (peer: Peer) => {
    setPeers((currentPeers) => {
      return  [...currentPeers,peer]
    });
  }


  useEffect(
    () => {
      var stream:WebSocket;
      if (track) {
        setPeers([])
        stream = ApiService.stream<PeerEvent>({
          route: "search_peers",
          onComplete: function () {
            setTrack(false)
          },
          onData: (resp)  => {
            if (resp.data.EventName === "sync") {
              addPeer(resp.data.Peer)
            } else if (resp.data.EventName === "update") {
              updatePeer(resp.data.Peer)
            } else if (resp.data.EventName === "delete") {
              deletePeer(resp.data.Peer)
            }
          }
        })
      }

      return () => {
        if(stream) {
          console.log("closing stream")
          stream.close()
        }
      }
    },
    [track],
  );

  const trackDmesh = () => {
    return (
      <span>
        <Icon type="play-circle" style={{marginRight: "10px"}}/>
      </span>
    )
  }

  const untrackDmesh = () => {
    return (
      <span>
        <Icon type="stop" style={{marginRight: "10px"}}/>
      </span>
    )
  }

  return (
    <MainLayout config={appConfig} {...props}>

      <PageHeader
        ghost={true}
        title="Dmesh Peers"
        extra={[
          <Btn key={1}  stopText={'Stop Tracking Dmesh'} startText={'Track Dmesh'} loading={track} onStart={() => setTrack(true)} onStop={() => setTrack(false)} />,
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Watch Key">
            {
              appConfig &&
              <Text code>/{appConfig.namespace}/{appConfig.dmeshServiceVersion}/search</Text>
            }
          </Descriptions.Item>
        </Descriptions>
      </PageHeader>
      <Row>
        <Col>
          {
            <React.Fragment>
              <div style={{marginTop: "10px"}}>
                <SearchPeerList peers={peers}/>
              </div>
            </React.Fragment>
          }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const DmeshPage = withRouter(BaseDmeshPage)