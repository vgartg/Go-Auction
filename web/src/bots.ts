import { currentUserID } from './auth';
import { placeBid } from './engine';
import { store } from './store';

const TICK_MS = 4500;

let timer: ReturnType<typeof setInterval> | null = null;

function step(): void {
  const me = currentUserID();
  const candidates = store.listLots('active').filter((l) => {
    const msLeft = new Date(l.closing_at).getTime() - Date.now();
    return msLeft > 0 && msLeft < 1000 * 60 * 60;
  });
  if (candidates.length === 0) return;
  if (Math.random() > 0.45) return;

  const lot = candidates[Math.floor(Math.random() * candidates.length)];
  const others = store.listUsers().filter((u) => u.id !== me);
  if (others.length === 0) return;

  const u = others[Math.floor(Math.random() * others.length)];
  const bumpSteps = 1 + Math.floor(Math.random() * 2);
  const amount = +(lot.current_price + lot.min_step * bumpSteps).toFixed(2);
  try {
    placeBid(lot.id, u.id, amount);
  } catch {
    /* ignore validation failures */
  }
}

export function startBots(): void {
  if (timer) return;
  timer = setInterval(step, TICK_MS);
}

export function stopBots(): void {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}
