import * as React from "react";
import styled from "styled-components/macro"
import { formatNumberWithCommas } from "../utils/format"

const BlockNumWrapper = styled.span`
  font-weight: bold;
  font-style: italic;
`

export const BlockNumRange: React.FC<{
  startBlockNum: number,
  endBlockNum: number,
  inv?: boolean,
}> = (props) => (
  <span className='black-range-num'>
    {
      props.inv &&
      (
        <>
          [<BlockNumWrapper>{formatNumberWithCommas(props.endBlockNum)}</BlockNumWrapper> - <BlockNumWrapper>{formatNumberWithCommas(props.startBlockNum)}</BlockNumWrapper>]
        </>
      )
    }
    {

      !props.inv &&
      (
        <>
          [<BlockNumWrapper>{formatNumberWithCommas(props.startBlockNum)}</BlockNumWrapper> - <BlockNumWrapper>{formatNumberWithCommas(props.endBlockNum)}</BlockNumWrapper>]
        </>
      )
    }
  </span>

)