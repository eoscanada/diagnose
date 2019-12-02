import React from "react"
import { Menu} from 'antd';
import { Link } from "react-router-dom";
import { Paths } from '../router/paths';

export function Navigation(props: {
  currentPath: string
}): React.ReactElement {
  return (
    <div>
      <Menu
        mode="inline"
        defaultSelectedKeys={[props.currentPath]}
        style={{ height: '100%' }}
      >
        <Menu.Item key={Paths.blocks}>
          <Link to={Paths.blocks}>Block Logs</Link>
        </Menu.Item>
        <Menu.Item key={Paths.indexes}>
          <Link to={Paths.indexes}>Search Indexes</Link>
        </Menu.Item>
        <Menu.Item key={Paths.dmesh}>
          <Link to={Paths.dmesh}>Dmesh Network</Link>
        </Menu.Item>
      </Menu>
    </div>
  )
}


