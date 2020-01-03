import React, { useEffect, useState } from "react"
import { Row, Col, Typography, PageHeader, Descriptions, Tag, Icon, Button } from "antd"
import { Peer } from "../types"
import { ApiService } from "../utils/api"
import { SearchPeerList } from "../components/search-peer-list"
import { MainLayout } from "../components/main-layout"
import { useStore } from "../store"

const { Text } = Typography

export const DmeshPage: React.FC = () => {
  const [visualize, setVisualize] = useState(false)
  const [connected, setConnected] = useState(false)
  const [peers, setPeers] = useState<Peer[]>([])
  const [headBlockNum, setHeadBlockNum] = useState(0)

  const [{ config: appConfig }] = useStore()

  useEffect(() => {
    setPeers([])
    setConnected(true)

    const deletePeer = (peer: Peer) => {
      setPeers((currentPeers) => {
        const newCurrentPeers = currentPeers.map((peerItem) => {
          if (peerItem.key === peer.key) {
            return { ...peerItem, deleted: true, ready: false }
          }
          return peerItem
        })
        return newCurrentPeers
      })
    }

    const updatePeer = (peer: Peer) => {
      if (peer.headBlockNum > headBlockNum) {
        setHeadBlockNum(peer.headBlockNum)
      }

      setPeers((currentPeers) => {
        let foundPeer = false
        let newCurrentPeers = currentPeers.map((peerItem) => {
          if (peerItem.key === peer.key) {
            foundPeer = true
            return peer
          }
          return peerItem
        })
        if (!foundPeer) {
          newCurrentPeers = [...newCurrentPeers, peer]
        }
        return newCurrentPeers
      })
    }

    const addPeer = (peer: Peer) => {
      if (peer.headBlockNum > headBlockNum) {
        setHeadBlockNum(peer.headBlockNum)
      }

      setPeers((currentPeers) => {
        return [...currentPeers, peer]
      })
    }

    const stream = ApiService.stream({
      route: "search_peers",
      onError: () => {
        setConnected(false)
      },
      onComplete: () => {
        setConnected(false)
      },
      onData: (resp) => {
        let localPeer
        switch (resp.type) {
          case "Transaction":
            break
          case "BlockRange":
            break
          case "Message":
            break
          case "PeerEvent":
            localPeer = resp.payload.Peer
            localPeer.key = resp.payload.PeerKey
            if (resp.payload.EventName === "sync") {
              console.log(`[SYNC] for peer ${resp.payload.PeerKey} - ${resp.payload.Peer.host}`)
              addPeer(localPeer)
            } else if (resp.payload.EventName === "update") {
              updatePeer(localPeer)
            } else if (resp.payload.EventName === "delete") {
              console.log(`[DELETED] for peer ${resp.payload.PeerKey} - ${resp.payload.Peer.host}`)
              deletePeer(localPeer)
            }
            break
        }
      }
    })

    return () => {
      if (stream) {
        stream.close()
      }
    }
  }, [headBlockNum])

  return (
    <MainLayout>
      <PageHeader
        title="Dmesh Peers"
        tags={
          connected ? (
            <Tag color="geekblue">
              <Icon type="sync" spin /> connected
            </Tag>
          ) : (
            <Tag color="red">
              <Icon type="disconnect" /> disconnected
            </Tag>
          )
        }
        extra={[
          <Button
            key="1"
            onClick={(e) => {
              e.preventDefault()
              setVisualize(!visualize)
            }}
            size="small"
            type="ghost"
          >
            <Icon type="sliders" />
          </Button>
        ]}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="Watch Key">
            {appConfig && appConfig.dmeshServiceVersion && (
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
            <SearchPeerList peers={peers} headBlockNum={headBlockNum} visualize={visualize} />
          </div>
        </Col>
      </Row>
    </MainLayout>
  )
}
