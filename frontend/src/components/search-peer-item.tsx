import React from "react"
import { Peer } from "../types"
import { Tag } from 'antd';
import { BlockNum } from "../atoms/block-num"
export function SearchPeerItem(props: {
  peer: Peer
}
): React.ReactElement {

  return (
    <tr>
      <td>{props.peer.host}</td>
      <td><Tag>{props.peer.tier}</Tag></td>
      <td><BlockNum blockNum={props.peer.firstBlockNum} /></td>
      <td><BlockNum blockNum={props.peer.irrBlockNum} /></td>
      <td><BlockNum blockNum={props.peer.headBlockNum} /></td>
      <td>{props.peer.shardSize}</td>
      <td>
        { props.peer.ready && <Tag color="#87d068">ready</Tag> }
        { !props.peer.ready && <Tag color="#f50">not ready</Tag> }
      </td>
      <td>
        { props.peer.reversible && <Tag color="#108ee9">Live</Tag> }
        { !props.peer.reversible && <Tag>Archive</Tag> }
      </td>
    </tr>
  )
}



