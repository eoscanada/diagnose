import React from "react"
import styled from "styled-components/macro"
import { Button } from "antd"

type Props = {
  startText: string,
  stopText: string,
  key?: any
  onStart?: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void
  onStop?: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void
  loading?: boolean
}

export function Btn(props: Props): React.ReactElement {
  const {
    startText,
    stopText,
    key,
    loading = false,
    onStart,
    onStop,
  } = props

  if (loading) {
    return (
      <Button key={key} type={"danger"} icon={"stop"}  onClick={onStop} >
        {stopText}
      </Button>
    )
  } else {
    return (
      <Button key={key} type={"primary"} icon={"play-circle"}  onClick={onStart} >
        {startText}
      </Button>
    )
  }
}


