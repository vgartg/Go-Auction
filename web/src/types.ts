export type LotStatus = 'active' | 'closed' | 'canceled';

export interface Lot {
  id: string;
  title: string;
  start_price: number;
  min_step: number;
  current_price: number;
  status: LotStatus;
  created_at: string;
  closing_at: string;
  version: number;
  winner_id: string | null;
  extended_count: number;
}

export interface Bid {
  id: string;
  lot_id: string;
  user_id: string;
  amount: number;
  created_at: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  password_hash: string;
  created_at: string;
}

export interface UserStats {
  user_id: string;
  username: string;
  bids_count: number;
  wins_count: number;
  total_spent: number;
}
