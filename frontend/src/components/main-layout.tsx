import React, { useState } from 'react';
import { DiagnoseConfig } from '../types'
import { TopHeader } from './top-header'
import { Navigation } from './navigation'
import { Layout } from 'antd';
const { Header, Footer, Sider, Content } = Layout;

export function MainLayout(props: {
  config : DiagnoseConfig | undefined,
  children: React.ReactNode
}): React.ReactElement {

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
          <Navigation />
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
