import React, { useState } from "react"
import { withRouter, RouteComponentProps } from "react-router"
import { PeerEvent, Peer} from "../types"
import {  ApiService } from "../utils/api";
import { useAppConfig } from "../hooks/dignose"
import { SearchPeerList } from "../components/search-peer-list"
import { MainLayout } from "../components/main-layout"
import {Typography, Row, Col, Button, Icon, List} from "antd"
const { Text } = Typography;

function BaseDmeshPage(
  props: RouteComponentProps
): React.ReactElement {

  const [processing, setProcessing] = useState(false)
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

  const updateDmesh = () => {
    setProcessing(true)
    const stream = ApiService.stream<PeerEvent>({
      route: "search_peers",
      onComplete: function () {
        setProcessing(false)
        console.log("search_peers completed")
      },
      onData: (resp)  => {
        console.log("resp: " ,resp)
        if (resp.data.EventName === "sync") {
          addPeer(resp.data.Peer)
        } else if (resp.data.EventName === "update") {
          updatePeer(resp.data.Peer)
        } else if (resp.data.EventName === "delete") {
          deletePeer(resp.data.Peer)
        }
      }
    })
    return () => {
      stream.close()
    }
  };

  return (
    <MainLayout config={appConfig}>
      <Row justify="space-between">
        <Col span={12} style={{ textAlign: "left"}}>
          <h1>dmesh</h1>
        </Col>
        <Col span={12} style={{ textAlign: "right"}}>
          <Button type="primary" loading={processing} onClick={updateDmesh}>
            refresh
            <Icon type="monitor" />
          </Button>
        </Col>
      </Row>
      <Row>
        <Col>
          {
            appConfig &&
            <SearchPeerList peers={peers}/>
          }
        </Col>
      </Row>
    </MainLayout>
  )
}

export const DmeshPage = withRouter(BaseDmeshPage)