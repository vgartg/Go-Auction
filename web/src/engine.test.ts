import { beforeEach, describe, expect, it } from 'vitest';
import { createLot, placeBid, SNIPING_WINDOW_MS } from './engine';
import { store } from './store';
import { uuid } from './uuid';

beforeEach(() => {
  store.reset();
});

describe('engine', () => {
  it('rejects bid below current price + min step', () => {
    const lot = createLot({
      title: 'Test',
      start_price: 100,
      min_step: 10,
      closing_at: new Date(Date.now() + 60_000),
    });
    const userID = uuid();
    expect(() => placeBid(lot.id, userID, 105)).toThrow();
  });

  it('accepts a valid bid and bumps current price', () => {
    const lot = createLot({
      title: 'Test',
      start_price: 100,
      min_step: 10,
      closing_at: new Date(Date.now() + 60_000),
    });
    const userID = uuid();
    const updated = placeBid(lot.id, userID, 110);
    expect(updated.current_price).toBe(110);
    expect(updated.version).toBe(2);
  });

  it('extends closing_at when bidding inside the anti-sniping window', () => {
    const lot = createLot({
      title: 'Test',
      start_price: 100,
      min_step: 10,
      closing_at: new Date(Date.now() + SNIPING_WINDOW_MS - 1000),
    });
    const originalClose = lot.closing_at;
    placeBid(lot.id, uuid(), 110);
    const after = store.getLot(lot.id)!;
    expect(after.extended_count).toBe(1);
    expect(new Date(after.closing_at).getTime()).toBeGreaterThan(new Date(originalClose).getTime());
  });

  it('refuses bids on closed lots', () => {
    const lot = createLot({
      title: 'Test',
      start_price: 100,
      min_step: 10,
      closing_at: new Date(Date.now() + 60_000),
    });
    lot.status = 'closed';
    store.upsertLot(lot);
    expect(() => placeBid(lot.id, uuid(), 200)).toThrow();
  });
});
