import React from "react"
import { Tag, Slider } from "antd"
import { Peer } from "../types"
import { BlockNum } from "../atoms/block-num"
import { formatNumberWithCommas } from "../utils/format"
import Moment from "react-moment"
import "moment-timezone"

type Props = {
  peer: Peer
  headBlockNum: number
  visualize: boolean
  showOther: boolean
}

export const SearchPeerItem: React.FC<Props> = ({ peer, headBlockNum, visualize, showOther }) => {
  let adjustedLowBlockNum = peer.tailBlockNum
  if (!peer.tailBlockNum) {
    adjustedLowBlockNum = 0
  }

  return (
    <>
      <tr className={peer.deleted ? "peer-deleted" : ""}>
        <td>{showOther ? peer.key : peer.host}</td>
        <td>
          <Tag>{peer.tier}</Tag>
        </td>
        {!visualize && (
          <>
            <td style={{ textAlign: "right" }}>
              {showOther ? peer.tailBlockID : <BlockNum blockNum={peer.tailBlockNum} />}
            </td>
            <td style={{ textAlign: "right" }}>
              {showOther ? peer.irrBlockID : <BlockNum blockNum={peer.irrBlockNum} />}
            </td>
            <td style={{ textAlign: "right" }}>
              {showOther ? peer.headBlockID : <BlockNum blockNum={peer.headBlockNum} />}
            </td>
            <td style={{ textAlign: "center" }}>{formatNumberWithCommas(peer.shardSize)}</td>
            <td style={{ textAlign: "center" }}>
              {peer.ready && <Tag color="green">ready</Tag>}
              {!peer.ready && peer.deleted && <Tag color="volcano">deleted</Tag>}
              {!peer.ready && !peer.deleted && <Tag color="magenta">not ready</Tag>}
            </td>
            <td style={{ textAlign: "center" }}>
              <Moment format="YYYY-MM-DD HH:mm Z">{peer.boot}</Moment>
              <br />
              <Moment fromNow>{peer.boot}</Moment>
            </td>
          </>
        )}
        {visualize && (
          <>
            <td>
              <Slider
                range
                value={[adjustedLowBlockNum, peer.headBlockNum]}
                min={85000000}
                max={headBlockNum}
              />
            </td>
          </>
        )}
        <td
          style={{
            textAlign: "center"
          }}
        >
          {peer.reversible && <Tag color={"#108ee9"}>Live</Tag>}
          {peer.headMoves && <Tag>Moving Head</Tag>}
          {peer.tailMoves && <Tag>Moving Tail</Tag>}
        </td>
      </tr>
    </>
  )
}
