export interface Transfer {
  id: string;
  amount: string;
  asset: string;
  from: string;
  to: string;
  timestamp: string;
  transactionHash: string;
}

export interface TransfersFilter {
  from?: string;
  to?: string;
  asset?: string;
  minAmount?: string;
  maxAmount?: string;
  startTime?: string;
  endTime?: string;
}

export interface TransfersSort {
  orderBy?: 'timestamp' | 'amount';
  orderDirection?: 'asc' | 'desc';
}
