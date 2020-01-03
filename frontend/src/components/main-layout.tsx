import React from "react"
import { useLocation } from "react-router"
import { Layout } from "antd"
import { TopHeader } from "./top-header"
import { Navigation } from "./navigation"

const { Header, Sider, Content } = Layout

export const MainLayout: React.FC = ({ children }) => {
  const { pathname } = useLocation()

  return (
    <Layout style={{ height: "100vh" }}>
      <Header
        style={{
          zIndex: 10,
          height: "70px",
          background: "#fff",
          boxShadow: "0 2px 8px #f0f1f2"
        }}
      >
        <TopHeader />
      </Header>
      <Layout
        style={{
          backgroundColor: "#fff"
        }}
      >
        <Sider
          style={{
            background: "#fff",
            boxShadow: "0 2px 8px #f0f1f2"
          }}
        >
          <Navigation currentPath={pathname} />
        </Sider>
        <Content
          style={{
            background: "#fff",
            padding: "50px"
          }}
        >
          {children}
        </Content>
      </Layout>
    </Layout>
  )
}
