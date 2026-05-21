import type { Bid, Lot } from './types';

export type LotEvent =
  | { type: 'new_bid'; lot_id: string; bid: Bid; new_price: number }
  | { type: 'lot_extended'; lot_id: string; closing_at: string; extended_count: number }
  | { type: 'lot_closed'; lot_id: string; winner_id: string | null; final_price: number }
  | { type: 'lot_changed'; lot: Lot };

export const bus = new EventTarget();

export function emit(ev: LotEvent): void {
  bus.dispatchEvent(new CustomEvent('lot', { detail: ev }));
}

export function onLot(handler: (ev: LotEvent) => void): () => void {
  const listener = (e: Event) => handler((e as CustomEvent<LotEvent>).detail);
  bus.addEventListener('lot', listener);
  return () => bus.removeEventListener('lot', listener);
}
