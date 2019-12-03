import React from "react";
import { Peer } from "../types";
import { SearchPeerItem } from "./search-peer-item";

export function SearchPeerList(props: { peers: Peer[] }): React.ReactElement {
  const peerItem = props.peers
    .sort((a: Peer, b: Peer) => {
      return a.tier < b.tier ? -1 : 1;
    })
    .map((peer: Peer) => <SearchPeerItem peer={peer} />);

  return (
    <div>
      <div className="ant-table-body">
        <table
          style={{
            width: "100%"
          }}
        >
          <thead className="ant-table-thead">
            <tr>
              <th>Host</th>
              <th>Tier</th>
              <th
                style={{
                  textAlign: "right"
                }}
              >
                Tail Block
              </th>
              <th
                style={{
                  textAlign: "right"
                }}
              >
                IRR Block
              </th>
              <th
                style={{
                  textAlign: "right"
                }}
              >
                Head Block
              </th>
              <th
                style={{
                  textAlign: "center"
                }}
              >
                Shard Size
              </th>
              <th
                style={{
                  textAlign: "center"
                }}
              >
                Status
              </th>
              <th
                style={{
                  textAlign: "center"
                }}
              >
                Reversible
              </th>
              <th
                style={{
                  textAlign: "center"
                }}
              >
                Moving Tail
              </th>
              <th
                style={{
                  textAlign: "center"
                }}
              >
                Moving Head
              </th>
            </tr>
          </thead>
          <tbody className="ant-table-tbody">{peerItem}</tbody>
        </table>
      </div>
    </div>
  );
}
