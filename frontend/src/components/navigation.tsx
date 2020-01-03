import React from "react"
import { Menu } from "antd"
import { Link } from "react-router-dom"
import { Paths } from "../router/paths"

export const Navigation: React.FC<{ currentPath: string }> = ({ currentPath }) => {
  return (
    <div>
      <Menu mode="inline" defaultSelectedKeys={[currentPath]} style={{ height: "100%" }}>
        <Menu.Item key={Paths.blocks}>
          <Link to={Paths.blocks}>Block Logs</Link>
        </Menu.Item>
        <Menu.Item key={Paths.indexes}>
          <Link to={Paths.indexes}>Search Indexes</Link>
        </Menu.Item>
        <Menu.Item key={Paths.kvdbBlocks}>
          <Link to={Paths.kvdbBlocks}>KVDB Blocks</Link>
        </Menu.Item>
        <Menu.Item key={Paths.kvdbTrx}>
          <Link to={Paths.kvdbTrx}>KVDB Trxs</Link>
        </Menu.Item>
        <Menu.Item key={Paths.dmesh}>
          <Link to={Paths.dmesh}>Dmesh Network</Link>
        </Menu.Item>
      </Menu>
    </div>
  )
}
