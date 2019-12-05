import React, { useState } from "react";
import { Icon, Tag, Slider } from "antd";
import { Peer } from "../types";
import { SearchPeerItem } from "./search-peer-item";

export function SearchPeerList(props: {
  peers: Peer[];
  headBlockNum: number;
}): React.ReactElement {
  const [visualize, setVisualize] = useState(false);
  const peerItem = props.peers
    .sort((a: Peer, b: Peer) => {
      return a.tier < b.tier ? -1 : 1;
    })
    .map((peer: Peer) => (
      <SearchPeerItem
        peer={peer}
        headBlockNum={props.headBlockNum}
        visualize={visualize}
      />
    ));

  return (
    <div>
      <div className="ant-table-body">
        <button
          style={{ float: "right" }}
          onClick={() => {
            setVisualize(!visualize);
          }}
        >
          <Icon type="radar-chart" />
        </button>
        <table style={{ width: "100%" }}>
          <thead className="ant-table-thead">
            <tr>
              <th>Host</th>
              <th>Tier</th>
              {!visualize && (
                <>
                  <th style={{ textAlign: "right" }}>Tail Block</th>
                  <th style={{ textAlign: "right" }}>IRR Block</th>
                  <th style={{ textAlign: "right" }}>Head Block</th>
                  <th style={{ textAlign: "center" }}>Shard Size</th>
                  <th style={{ textAlign: "center" }}>Status</th>
                  <th style={{ textAlign: "center" }}>Boot Time</th>
                </>
              )}
              {visualize && (
                <>
                  <th style={{ width: "100%" }}></th>
                </>
              )}
              <th style={{ textAlign: "center" }}>Features</th>
            </tr>
          </thead>
          <tbody className="ant-table-tbody">{peerItem}</tbody>
        </table>
      </div>
    </div>
  );
}
