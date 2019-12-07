import React, { useState } from "react";
import { Icon, Button } from "antd";
import { Peer } from "../types";
import { SearchPeerItem } from "./search-peer-item";

export function SearchPeerList(props: {
  peers: Peer[];
  visualize: boolean;
  headBlockNum: number;
}): React.ReactElement {
  const [showKey, setShowKey] = useState(false);

  const peerItem = props.peers
    .sort((a: Peer, b: Peer) => {
      return a.tier < b.tier ? -1 : 1;
    })
    .map((peer: Peer) => (
      <SearchPeerItem
        key={`${peer.host}-${peer.key}`}
        peer={peer}
        headBlockNum={props.headBlockNum}
        visualize={props.visualize}
        showKey={showKey}
      />
    ));

  return (
    <div>
      <div className="ant-table-body">
        <table style={{ width: "100%" }}>
          <thead className="ant-table-thead">
            <tr>
              <th>
                <Icon type="key" onClick={() => setShowKey(!showKey)} /> Host
              </th>
              <th>Tier</th>
              {!props.visualize && (
                <>
                  <th style={{ textAlign: "right" }}>Tail Block</th>
                  <th style={{ textAlign: "right" }}>IRR Block</th>
                  <th style={{ textAlign: "right" }}>Head Block</th>
                  <th style={{ textAlign: "center" }}>Shard Size</th>
                  <th style={{ textAlign: "center" }}>Status</th>
                  <th style={{ textAlign: "center" }}>Boot Time</th>
                </>
              )}
              {props.visualize && <th style={{ width: "100%" }}></th>}
              <>
                <th style={{ textAlign: "center" }}>Features</th>
                <th></th>
              </>
            </tr>
          </thead>
          <tbody className="ant-table-tbody">{peerItem}</tbody>
        </table>
      </div>
    </div>
  );
}
