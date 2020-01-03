import React from "react"
import { Icon, List } from "antd"
import { BlockRange } from "../types"
import { BlockNumRange } from "../atoms/block-num-range"
import { BlockNum } from "../atoms/block-num"
import { formatNanoseconds } from "../utils/format"

type Props = {
  ranges: BlockRange[]
  inv?: boolean
  elapsed?: number
}

export const BlockHolesList: React.FC<Props> = ({ ranges, inv, elapsed }) => {
  let header = <div />

  if (ranges.length > 0) {
    if (inv) {
      header = (
        <div>
          Start block: <BlockNum blockNum={ranges[0].endBlock} />
          <span
            style={{
              float: "right"
            }}
          >
            elapsed: {formatNanoseconds(elapsed || 0)}
          </span>
        </div>
      )
    } else {
      header = (
        <div>
          Validated up to block log: <BlockNum blockNum={ranges[ranges.length - 1].endBlock} />
          <span
            style={{
              float: "right"
            }}
          >
            elapsed: {formatNanoseconds(elapsed || 0)}
          </span>
        </div>
      )
    }
  }

  const renderBlockRange = (range: BlockRange) => {
    return (
      <List.Item>
        <div className="block-range-data-item">
          {range.status === "valid" && (
            <Icon
              style={{ fontSize: "24px" }}
              type="check-circle"
              theme="twoTone"
              twoToneColor="#52c41a"
            />
          )}
          {range.status === "hole" && (
            <Icon
              style={{ fontSize: "24px" }}
              type="close-circle"
              theme="twoTone"
              twoToneColor="#f5222d"
            />
          )}

          <BlockNumRange startBlockNum={range.startBlock} endBlockNum={range.endBlock} inv={inv} />
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
        dataSource={ranges}
        renderItem={(item) => renderBlockRange(item)}
      />
    </div>
  )
}
