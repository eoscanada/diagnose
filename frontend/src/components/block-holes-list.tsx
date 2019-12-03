import React from "react";
import { Icon, List } from "antd";
import { BlockRange } from "../types";
import { BlockNumRange } from "../atoms/block-num-range";
import { BlockNum } from "../atoms/block-num";
import { formatNanoseconds } from "../utils/format";

export function BlockHolesList(props: {
  ranges: BlockRange[];
  inv?: boolean;
  elapsed?: number;
}): React.ReactElement {
  let header = <div />;

  if (props.ranges.length > 0) {
    if (props.inv) {
      header = (
        <div>
          Start block: <BlockNum blockNum={props.ranges[0].endBlock} />
          <span
            style={{
              float: "right"
            }}
          >
            elapsed: {formatNanoseconds(props.elapsed || 0)}
          </span>
        </div>
      );
    } else {
      header = (
        <div>
          Validated up to block log:{" "}
          <BlockNum blockNum={props.ranges[props.ranges.length - 1].endBlock} />
          <span
            style={{
              float: "right"
            }}
          >
            elapsed: {formatNanoseconds(props.elapsed || 0)}
          </span>
        </div>
      );
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

          <BlockNumRange
            startBlockNum={range.startBlock}
            endBlockNum={range.endBlock}
            inv={props.inv}
          />
          {range.message}
        </div>
      </List.Item>
    );
  };

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
  );
}
