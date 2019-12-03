import React from "react";
import { Tag } from "antd";
import { Peer } from "../types";
import { BlockNum } from "../atoms/block-num";
import { formatNumberWithCommas } from "../utils/format";

export function SearchPeerItem(props: { peer: Peer }): React.ReactElement {
  return (
    <tr>
      <td>{props.peer.host}</td>
      <td>
        <Tag>{props.peer.tier}</Tag>
      </td>
      <td
        style={{
          textAlign: "right"
        }}
      >
        <BlockNum blockNum={props.peer.firstBlockNum} />
      </td>
      <td
        style={{
          textAlign: "right"
        }}
      >
        <BlockNum blockNum={props.peer.irrBlockNum} />
      </td>
      <td
        style={{
          textAlign: "right"
        }}
      >
        <BlockNum blockNum={props.peer.headBlockNum} />
      </td>
      <td
        style={{
          textAlign: "center"
        }}
      >
        {formatNumberWithCommas(props.peer.shardSize)}
      </td>
      <td
        style={{
          textAlign: "center"
        }}
      >
        {props.peer.ready && <Tag color="#87d068">ready</Tag>}
        {!props.peer.ready && <Tag color="#f50">not ready</Tag>}
      </td>
      <td
        style={{
          textAlign: "center"
        }}
      >
        {props.peer.tailMoves && <Tag color="#108ee9">Enabled</Tag>}
        {!props.peer.tailMoves && <Tag>Disabled</Tag>}
      </td>
      <td
        style={{
          textAlign: "center"
        }}
      >
        {props.peer.reversible && <Tag color="#108ee9">Live</Tag>}
        {!props.peer.reversible && <Tag>Archive</Tag>}
      </td>
      <td
        style={{
          textAlign: "center"
        }}
      >
        {props.peer.headMoves && <Tag color="#108ee9">Enabled</Tag>}
        {!props.peer.headMoves && <Tag>Disabled</Tag>}
      </td>
    </tr>
  );
}
