import React, { useState } from "react"
import { Icon } from "antd"
import { Peer } from "../types"
import { SearchPeerItem } from "./search-peer-item"

type Props = {
  peers: Peer[]
  visualize: boolean
  headBlockNum: number
}

export const SearchPeerList: React.FC<Props> = ({ peers, headBlockNum, visualize }) => {
  const [showOther, setShowOther] = useState(false)

  const peerItem = peers
    .sort((a: Peer, b: Peer) => {
      if (a.tier === b.tier) {
        return a.host < b.host ? -1 : 1
      }
      return a.tier < b.tier ? -1 : 1
    })
    .map((peer: Peer) => (
      <SearchPeerItem
        key={`${peer.host}-${peer.key}`}
        peer={peer}
        headBlockNum={headBlockNum}
        visualize={visualize}
        showOther={showOther}
      />
    ))

  return (
    <div>
      <div className="ant-table-body">
        <table style={{ width: "100%" }}>
          <thead className="ant-table-thead">
            <tr>
              <th>
                <Icon type="key" onClick={() => setShowOther(!showOther)} />{" "}
                {showOther ? "Addr" : "Host"}
              </th>
              <th>Tier</th>
              {!visualize && (
                <>
                  <th style={{ textAlign: "right" }}>{showOther ? "Tail Id" : "Tail Block"}</th>
                  <th style={{ textAlign: "right" }}>{showOther ? "IRR Id" : "IRR Block"}</th>
                  <th style={{ textAlign: "right" }}>{showOther ? "Head Id" : "Head Block"}</th>
                  <th style={{ textAlign: "center" }}>Shard Size</th>
                  <th style={{ textAlign: "center" }}>Status</th>
                  <th style={{ textAlign: "center" }}>Boot Time</th>
                </>
              )}
              {visualize && <th style={{ width: "100%" }}></th>}
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
  )
}
