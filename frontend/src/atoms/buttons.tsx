import React from "react"
import styled from "styled-components/macro"
import { Button } from "antd"

type Props = {
  children: React.ReactNode
  type?: "default"| "primary"| "ghost" | "dashed" | "danger" | "link",
  key?: any
  onClick?: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void
  loading?: boolean
  icon?: React.ReactNode
}

export function Btn(props: Props): React.ReactElement {
  const {
    children,
    key,
    icon,
    type = 'primary',
    loading = false,
    onClick,
  } = props

  return (
    <Button
      className={'btn'}
      key={key}
      type={type}
      onClick={onClick}
      loading={loading}
    >
      <span className={'svg'}>
        {icon}
      </span>
      {children}
    </Button>
  )
}

