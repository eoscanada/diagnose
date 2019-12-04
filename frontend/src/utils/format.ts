import numeral from "numeral";

export function formatNumberWithCommas(input: number | string): string {
  return numeral(input).format("0,0");
}

export function formatNanoseconds(nano: number): string {
  return `${nano / 1000000000.0} s`;
}

export function formatTime(boot: string): string {
  return boot;
}
