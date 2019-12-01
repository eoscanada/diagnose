import React from "react"
import { Peer } from "../types"
import { SearchPeerItem } from "./search-peer-item";

export function SearchPeerList(props: {
    peers: Peer[]
  }
): React.ReactElement {

  const peerItem = props.peers.map((peer) =>
    <SearchPeerItem peer={peer} />
  );


  return (
    <div>
      <h1>Peer Collection</h1>
      <div className={"ant-table-body"}>
        <table style={{
          width: "100%"
        }}>
          <thead className={"ant-table-thead"}>
            <tr>
              <th>Host</th>
              <th>Tier</th>
              <th>First Block</th>
              <th>IRR Block</th>
              <th>Head Block</th>
              <th>Shard Size</th>
              <th>Status</th>
              <th>Reversible</th>
            </tr>
          </thead>
          <tbody className={"ant-table-tbody"}>
            {peerItem}
          </tbody>
        </table>
      </div>
    </div>
  )
}

