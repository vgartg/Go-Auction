import { createLot, placeBid, scheduleClose } from './engine';
import { store } from './store';
import type { Bid, User } from './types';
import { uuid } from './uuid';

const SEED_FLAG = 'goauction:seeded:v2';

function makeUser(username: string, email: string): User {
  return {
    id: uuid(),
    username,
    email,
    password_hash: 'demo$seed',
    created_at: new Date(Date.now() - 86_400_000).toISOString(),
  };
}

export function seedIfEmpty(): void {
  if (localStorage.getItem(SEED_FLAG)) {
    for (const lot of store.listLots('active')) scheduleClose(lot);
    return;
  }

  const alice = makeUser('alice', 'alice@example.com');
  const bob = makeUser('bob', 'bob@example.com');
  const carol = makeUser('carol', 'carol@example.com');
  store.addUser(alice);
  store.addUser(bob);
  store.addUser(carol);

  const now = Date.now();

  const lotNear = createLot({
    title: 'Vintage Leica M3 (1958)',
    start_price: 1200,
    min_step: 50,
    closing_at: new Date(now + 90_000),
  });
  placeBid(lotNear.id, alice.id, 1250);
  placeBid(lotNear.id, bob.id, 1350);
  placeBid(lotNear.id, carol.id, 1500);

  const lotFar = createLot({
    title: 'First edition Foundation by Asimov',
    start_price: 400,
    min_step: 20,
    closing_at: new Date(now + 1000 * 60 * 60 * 6),
  });
  placeBid(lotFar.id, alice.id, 420);
  placeBid(lotFar.id, bob.id, 460);

  const lotFresh = createLot({
    title: 'Mid-century walnut desk',
    start_price: 800,
    min_step: 25,
    closing_at: new Date(now + 1000 * 60 * 25),
  });
  placeBid(lotFresh.id, carol.id, 825);

  const closedID = uuid();
  store.upsertLot({
    id: closedID,
    title: 'Signed Wozniak poster',
    start_price: 200,
    min_step: 10,
    current_price: 360,
    status: 'closed',
    created_at: new Date(now - 1000 * 60 * 60 * 24 * 3).toISOString(),
    closing_at: new Date(now - 1000 * 60 * 60 * 2).toISOString(),
    version: 5,
    winner_id: bob.id,
    extended_count: 1,
  });
  const closedBids: Bid[] = [
    { id: uuid(), lot_id: closedID, user_id: alice.id, amount: 220, created_at: new Date(now - 1000 * 60 * 60 * 5).toISOString() },
    { id: uuid(), lot_id: closedID, user_id: carol.id, amount: 280, created_at: new Date(now - 1000 * 60 * 60 * 4).toISOString() },
    { id: uuid(), lot_id: closedID, user_id: bob.id, amount: 360, created_at: new Date(now - 1000 * 60 * 60 * 2 - 5_000).toISOString() },
  ];
  for (const b of closedBids) store.addBid(b);

  localStorage.setItem(SEED_FLAG, '1');
}
