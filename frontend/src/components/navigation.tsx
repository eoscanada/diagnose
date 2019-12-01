import React from "react"
import { Menu} from 'antd';
import { Link } from "react-router-dom";
import { Paths } from '../router/paths';

export function Navigation(): React.ReactElement {
  return (
    <div>
      <Menu
        mode="inline"
        defaultSelectedKeys={['1']}
        style={{ height: '100%' }}
      >
        <Menu.Item key="1">
          <Link to={Paths.blocks}>Block Logs</Link>
        </Menu.Item>
        <Menu.Item key="2">
          <Link to={Paths.indexes}>Search Indexes</Link>
        </Menu.Item>
        <Menu.Item key="3">
          <Link to={Paths.dmesh}>Dmesh Network</Link>
        </Menu.Item>
      </Menu>
    </div>
  )
}


