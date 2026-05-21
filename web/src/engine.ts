import { emit } from './events';
import { store } from './store';
import type { Bid, Lot } from './types';
import { uuid } from './uuid';

export const SNIPING_WINDOW_MS = 30_000;
export const SNIPING_EXTENSION_MS = 30_000;

const timers = new Map<string, ReturnType<typeof setTimeout>>();

export function createLot(input: {
  title: string;
  start_price: number;
  min_step: number;
  closing_at: Date;
}): Lot {
  if (!input.title.trim()) throw new Error('title is required');
  if (input.start_price <= 0) throw new Error('start price must be positive');
  if (input.min_step <= 0) throw new Error('min step must be positive');
  if (input.closing_at.getTime() <= Date.now()) throw new Error('closing time must be in the future');

  const lot: Lot = {
    id: uuid(),
    title: input.title.trim(),
    start_price: input.start_price,
    min_step: input.min_step,
    current_price: input.start_price,
    status: 'active',
    created_at: new Date().toISOString(),
    closing_at: input.closing_at.toISOString(),
    version: 1,
    winner_id: null,
    extended_count: 0,
  };
  store.upsertLot(lot);
  scheduleClose(lot);
  emit({ type: 'lot_changed', lot });
  return lot;
}

export function placeBid(lotID: string, userID: string, amount: number): Lot {
  const lot = store.getLot(lotID);
  if (!lot) throw new Error('lot not found');
  if (lot.status !== 'active') throw new Error('lot is not active');
  if (Date.now() > new Date(lot.closing_at).getTime()) throw new Error('lot already closed');
  if (amount <= lot.current_price) {
    throw new Error(`bid must be higher than current price ${lot.current_price.toFixed(2)}`);
  }
  if (amount < lot.current_price + lot.min_step) {
    throw new Error(`bid must be at least ${lot.min_step.toFixed(2)} more than current price`);
  }

  const bid: Bid = {
    id: uuid(),
    lot_id: lotID,
    user_id: userID,
    amount,
    created_at: new Date().toISOString(),
  };
  store.addBid(bid);

  lot.current_price = amount;
  lot.version += 1;

  let extended = false;
  const msToClose = new Date(lot.closing_at).getTime() - Date.now();
  if (msToClose <= SNIPING_WINDOW_MS) {
    lot.closing_at = new Date(Date.now() + SNIPING_EXTENSION_MS).toISOString();
    lot.extended_count += 1;
    extended = true;
  }
  store.upsertLot(lot);

  emit({ type: 'new_bid', lot_id: lotID, bid, new_price: amount });
  emit({ type: 'lot_changed', lot });
  if (extended) {
    scheduleClose(lot);
    emit({
      type: 'lot_extended',
      lot_id: lotID,
      closing_at: lot.closing_at,
      extended_count: lot.extended_count,
    });
  }
  return lot;
}

export function closeLot(lotID: string): void {
  const lot = store.getLot(lotID);
  if (!lot || lot.status !== 'active') return;
  if (Date.now() < new Date(lot.closing_at).getTime()) {
    scheduleClose(lot);
    return;
  }
  const top = store.highestBid(lotID);
  if (top) lot.winner_id = top.user_id;
  lot.status = 'closed';
  lot.version += 1;
  store.upsertLot(lot);
  clearTimer(lotID);
  emit({
    type: 'lot_closed',
    lot_id: lotID,
    winner_id: lot.winner_id,
    final_price: lot.current_price,
  });
  emit({ type: 'lot_changed', lot });
}

export function scheduleClose(lot: Lot): void {
  clearTimer(lot.id);
  const ms = new Date(lot.closing_at).getTime() - Date.now();
  if (ms <= 0) {
    queueMicrotask(() => closeLot(lot.id));
    return;
  }
  const t = setTimeout(() => closeLot(lot.id), Math.min(ms, 2_147_483_000));
  timers.set(lot.id, t);
}

function clearTimer(lotID: string): void {
  const t = timers.get(lotID);
  if (t) {
    clearTimeout(t);
    timers.delete(lotID);
  }
}

export function restoreTimers(): void {
  for (const lot of store.listLots('active')) scheduleClose(lot);
}
