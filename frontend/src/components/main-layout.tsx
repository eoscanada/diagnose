import React  from 'react';
import { RouteComponentProps } from "react-router";
import { Layout } from 'antd';
import { DiagnoseConfig } from '../types'
import { TopHeader } from './top-header'
import { Navigation } from './navigation'


const { Header, Sider, Content } = Layout;

export function MainLayout(
  props: {
    config : DiagnoseConfig | undefined,
    children: React.ReactNode
  } & RouteComponentProps
): React.ReactElement {
  return (
    <Layout style={{height:"100vh"}}>
      <Header
        style={{
          zIndex: 10,
          height: "70px",
          background: '#fff',
          boxShadow: "0 2px 8px #f0f1f2",
        }}>

        <TopHeader config={props.config} />
      </Header>
      <Layout
        style={{
          backgroundColor: '#fff',
        }}
      >
        <Sider
          style={{
            background: '#fff',
            boxShadow: "0 2px 8px #f0f1f2"
          }}>
          <Navigation currentPath={props.location.pathname} />
        </Sider>
        <Content
          style={{
            background: '#fff',
            padding: '50px'
          }}>
          {props.children}
        </Content>
      </Layout>
    </Layout>
  );
}
