import numeral from "numeral"


export function formatNumberWithCommas(input: number | string): string {
  let str = numeral(input).format("000000000")
  let formatStr = ""
  for (var i = 0; i <  str.length; i++) {
    if (i%3 === 0) {
      formatStr = formatStr + ","
    }
    formatStr = formatStr + str[i]
  }
  return formatStr.slice(1,12)
}



