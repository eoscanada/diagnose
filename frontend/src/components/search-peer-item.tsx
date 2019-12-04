import React from "react";
import { Tag } from "antd";
import { Peer } from "../types";
import { BlockNum } from "../atoms/block-num";
import { formatNumberWithCommas, formatTime } from "../utils/format";
import Moment from "react-moment";
import "moment-timezone";

export function SearchPeerItem(props: { peer: Peer }): React.ReactElement {
  return (
    <tr>
      <td>{props.peer.host}</td>
      <td>
        <Tag>{props.peer.tier}</Tag>
      </td>
      <td style={{ textAlign: "right" }}>
        <BlockNum blockNum={props.peer.tailBlockNum} />
      </td>
      <td style={{ textAlign: "right" }}>
        <BlockNum blockNum={props.peer.irrBlockNum} />
      </td>
      <td style={{ textAlign: "right" }}>
        <BlockNum blockNum={props.peer.headBlockNum} />
      </td>
      <td style={{ textAlign: "center" }}>
        {formatNumberWithCommas(props.peer.shardSize)}
      </td>
      <td style={{ textAlign: "center" }}>
        {props.peer.ready && <Tag color="#87d068">ready</Tag>}
        {!props.peer.ready && <Tag color="#f50">not ready</Tag>}
      </td>
      <td style={{ textAlign: "center" }}>
        <Moment format="YYYY-MM-DD HH:mm Z">{props.peer.boot}</Moment>
      </td>
      <td
        style={{
          textAlign: "center"
        }}
      >
        {props.peer.reversible && <Tag color={"#108ee9"}>Live</Tag>}
        {props.peer.headMoves && <Tag>Moving Head</Tag>}
        {props.peer.tailMoves && <Tag>Moving Tail</Tag>}
      </td>
    </tr>
  );
}
