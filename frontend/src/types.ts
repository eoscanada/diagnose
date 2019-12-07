export interface DiagnoseConfig {
  protocol: string;
  namespace: string;
  blockStoreUrl: string;
  indexesStoreUrl: string;
  shardSize: number;
  kvdbConnectionInfo: string;
  dmeshServiceVersion: string;
}

export type Progress = ProgressSocketMessage["payload"];
export interface ProgressSocketMessage {
  type: "Progress";
  payload: {
    elapsed: number;
    totalIteration: number;
    currentIteration: number;
  };
}

export type Transaction = TransactionSocketMessage["payload"];
export interface TransactionSocketMessage {
  type: "Transaction";
  payload: {
    prefix: string;
    id: string;
    blockNum: number;
  };
}

export type BlockRange = BlockRangeSocketMessage["payload"];
export interface BlockRangeSocketMessage {
  type: "BlockRange";
  payload: {
    startBlock: number;
    endBlock: number;
    message: string;
    status: "valid" | "hole";
  };
}

export type Message = MessageSocketMessage["payload"];
export interface MessageSocketMessage {
  type: "Message";
  payload: {
    message: string;
  };
}

export type PeerEvent = PeerEventSocketMessage["payload"];
export interface PeerEventSocketMessage {
  type: "PeerEvent";
  payload: {
    EventName: string;
    PeerKey: string;
    Peer: Peer;
  };
}

export interface Peer {
  boot: string;
  tailBlockNum: number;
  headBlockID: string;
  headBlockNum: number;
  headMoves: boolean;
  host: string;
  irrBlockID: string;
  irrBlockNum: number;
  ready: boolean;
  reversible: boolean;
  shardSize: number;
  tailMoves: boolean;
  tier: number;
  deleted: boolean;
  key: string;
}

export type SocketMessage =
  | TransactionSocketMessage
  | BlockRangeSocketMessage
  | MessageSocketMessage
  | PeerEventSocketMessage
  | ProgressSocketMessage;

export interface ApiResponse<T> {
  data: T;
  type: string;
  errors: string[] | undefined;
}
