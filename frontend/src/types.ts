export interface DiagnoseConfig {
  protocol: string
  namespace: string
  blockStoreUrl: string
  indexesStoreUrl: string
  shardSize: number
  kvdbConnectionInfo: string
}

export interface BlockRangeData {
  startBlock: number
  endBlock: number
  message: string
  status: "valid" | "hole"
}

export interface ApiResponse<T> {
  data: T,
  errors: string[] | undefined
}

export interface PeerEvent {
  EventName: string
  Peer: Peer
}

export interface  Peer {
  boot: string
  firstBlockNum: number
  headBlockID: string
  headBlockNum: number
  headMoves: boolean
  host: string
  irrBlockID: string
  irrBlockNum: number
  ready: boolean
  reversible: boolean
  shardSize: number
  tailMoves: boolean
  tier: number
  deleted: boolean
}