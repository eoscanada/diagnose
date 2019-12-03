import React from "react"
import { BlockRangeData } from "../types"
import { Icon, List } from "antd"
import { BlockNumRange } from "../atoms/block-num-range"
import { BlockNum } from "../atoms/block-num"

export function BlockHolesList(props: {
    ranges: BlockRangeData[],
    inv?: boolean,
  }
): React.ReactElement {


  let header = (<div></div>)

  if(props.ranges.length > 0) {
    if( props.inv) {
      header = (<div>Start block: {<BlockNum blockNum={props.ranges[0].endBlock} />}</div>)
    } else {
      header = (<div>Validated up to block log: {<BlockNum blockNum={props.ranges[props.ranges.length -1].endBlock} />}</div>)
    }
  }


  const renderBlockRange = (range: BlockRangeData) => {
    return (
      <List.Item>
        <div className={"block-range-data-item"}>
          {(range.status === "valid") && <Icon style={{fontSize: "24px"}} type="check-circle" theme="twoTone" twoToneColor="#52c41a"/>}
          {(range.status === "hole") && <Icon style={{fontSize: "24px"}}  type="close-circle" theme="twoTone" twoToneColor="#f5222d"/>}

          <BlockNumRange startBlockNum={range.startBlock} endBlockNum={range.endBlock} inv={props.inv}/>
          {range.message}
        </div>
      </List.Item>
    )
  }

  return (
    <div>
      <List
        size="small"
        header={header}
        bordered
        dataSource={props.ranges}
        renderItem={item => renderBlockRange(item)}
      />
    </div>

  )
}

