import type { Bid, Lot, User } from './types';

const KEY = 'goauction:state:v1';

interface State {
  lots: Record<string, Lot>;
  bids: Bid[];
  users: Record<string, User>;
}

function emptyState(): State {
  return { lots: {}, bids: [], users: {} };
}

let cache: State | null = null;

function load(): State {
  if (cache) return cache;
  try {
    const raw = localStorage.getItem(KEY);
    cache = raw ? (JSON.parse(raw) as State) : emptyState();
  } catch {
    cache = emptyState();
  }
  return cache;
}

function save(): void {
  if (!cache) return;
  localStorage.setItem(KEY, JSON.stringify(cache));
}

export const store = {
  reset(): void {
    cache = emptyState();
    save();
  },

  upsertLot(lot: Lot): void {
    const s = load();
    s.lots[lot.id] = lot;
    save();
  },

  getLot(id: string): Lot | null {
    const s = load();
    return s.lots[id] ?? null;
  },

  listLots(status?: string): Lot[] {
    const s = load();
    const all = Object.values(s.lots);
    const filtered = status ? all.filter((l) => l.status === status) : all;
    return filtered.sort((a, b) => (a.created_at < b.created_at ? 1 : -1));
  },

  addBid(bid: Bid): void {
    const s = load();
    s.bids.push(bid);
    save();
  },

  recentBids(lotID: string, limit: number): Bid[] {
    const s = load();
    return s.bids
      .filter((b) => b.lot_id === lotID)
      .sort((a, b) => (a.created_at < b.created_at ? 1 : -1))
      .slice(0, limit);
  },

  highestBid(lotID: string): Bid | null {
    const s = load();
    const lotBids = s.bids.filter((b) => b.lot_id === lotID);
    if (lotBids.length === 0) return null;
    return lotBids.reduce((best, b) => (b.amount > best.amount ? b : best));
  },

  addUser(user: User): void {
    const s = load();
    s.users[user.id] = user;
    save();
  },

  getUser(id: string): User | null {
    const s = load();
    return s.users[id] ?? null;
  },

  findUserByEmail(email: string): User | null {
    const s = load();
    return Object.values(s.users).find((u) => u.email.toLowerCase() === email.toLowerCase()) ?? null;
  },

  findUserByUsername(name: string): User | null {
    const s = load();
    return Object.values(s.users).find((u) => u.username.toLowerCase() === name.toLowerCase()) ?? null;
  },

  bidsByUser(userID: string): Bid[] {
    const s = load();
    return s.bids.filter((b) => b.user_id === userID);
  },

  winsByUser(userID: string): Lot[] {
    const s = load();
    return Object.values(s.lots).filter((l) => l.winner_id === userID);
  },

  listUsers(): User[] {
    const s = load();
    return Object.values(s.users);
  },
};
