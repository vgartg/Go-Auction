import { currentUser } from '../auth';
import { placeBid } from '../engine';
import { onLot } from '../events';
import { store } from '../store';
import type { Bid, Lot, User } from '../types';
import { bidRow, bidderLabel, statusBadge } from '../ui/components';
import { escape, render } from '../ui/layout';

let cleanup: (() => void) | null = null;
let timerHandle: ReturnType<typeof setInterval> | null = null;

function teardown(): void {
  if (cleanup) {
    cleanup();
    cleanup = null;
  }
  if (timerHandle) {
    clearInterval(timerHandle);
    timerHandle = null;
  }
}

function formatRemaining(ms: number): string {
  if (ms <= 0) return '0s';
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  const h = Math.floor(m / 60);
  const sec = s % 60;
  const min = m % 60;
  return `${h ? h + 'h ' : ''}${min ? min + 'm ' : ''}${sec}s`;
}

function bidPanel(lot: Lot, user: User | null, errorMsg = ''): string {
  if (!user) {
    return `<p class="text-sm text-slate-500"><a href="#/login" class="text-amber-600 hover:underline">Log in</a> to place a bid.</p>`;
  }
  if (lot.status !== 'active') {
    return `<p class="text-sm text-slate-500">This lot is no longer active.</p>`;
  }
  const minBid = (lot.current_price + lot.min_step).toFixed(2);
  return `
    <form id="bid-form" class="space-y-2">
      <label class="block text-xs text-slate-500">Your bid (min ${minBid})</label>
      <input
        id="bid-amount"
        type="number"
        name="amount"
        step="0.01"
        min="${minBid}"
        value="${minBid}"
        required
        class="w-full border border-slate-300 rounded-md px-3 py-2 text-base tabular-nums focus:outline-none focus:ring-2 focus:ring-amber-400"
      />
      ${errorMsg ? `<div class="text-xs text-red-600">${escape(errorMsg)}</div>` : ''}
      <button type="submit" class="w-full bg-amber-500 hover:bg-amber-600 text-white font-medium rounded-md px-3 py-2">Bid</button>
    </form>
  `;
}

function closedBlock(lot: Lot): string {
  const winnerLabel = lot.winner_id
    ? `<a href="#/users/${lot.winner_id}" class="text-amber-700 hover:underline text-sm">Winner: ${escape(bidderLabel(lot.winner_id))} →</a>`
    : `<span class="text-slate-500">No bids placed.</span>`;
  return `
    <div class="bg-slate-100 rounded-md p-4 text-sm">
      <div class="text-xs uppercase tracking-wide text-slate-500 mb-1">Final price</div>
      <div class="text-2xl font-bold mb-2">${lot.current_price.toFixed(2)}</div>
      ${winnerLabel}
    </div>
  `;
}

function activeBlock(lot: Lot): string {
  return `
    <div class="bg-white rounded-md border border-slate-200 p-4 mb-4">
      <div class="text-xs uppercase tracking-wide text-slate-500 mb-1">Closes in</div>
      <div id="time-remaining" class="text-2xl font-semibold tabular-nums" data-closes-at="${escape(lot.closing_at)}">—</div>
      <div id="extended-badge" class="mt-2 inline-block bg-amber-50 text-amber-700 text-xs font-medium rounded px-2 py-0.5 ${lot.extended_count === 0 ? 'hidden' : ''}">
        Anti-sniping: extended × <span id="extended-count">${lot.extended_count}</span>
      </div>
    </div>
  `;
}

function bidHistory(bids: Bid[]): string {
  if (bids.length === 0) {
    return `<li class="text-slate-400 py-2">No bids yet.</li>`;
  }
  return bids.map(bidRow).join('');
}

function renderLot(lotID: string, flash = ''): void {
  teardown();
  const lot = store.getLot(lotID);
  if (!lot) {
    render('Not found', `<div class="text-center text-slate-500 py-12">Lot not found.</div>`);
    return;
  }
  const user = currentUser();
  const bids = store.recentBids(lotID, 20);

  const body = `
    <div class="grid lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2 bg-white rounded-lg border border-slate-200 p-6">
        <div class="flex items-start justify-between mb-2">
          <h1 class="text-2xl font-bold">${escape(lot.title)}</h1>
          ${statusBadge(lot.status)}
        </div>
        <div class="text-xs text-slate-500 mb-6 font-mono">${escape(lot.id)}</div>

        <div class="bg-slate-50 rounded-md p-4 mb-4">
          <div class="text-xs uppercase tracking-wide text-slate-500 mb-1">Current price</div>
          <div id="current-price" class="text-4xl font-bold tabular-nums">${lot.current_price.toFixed(2)}</div>
          <div class="text-xs text-slate-500 mt-2">Start ${lot.start_price.toFixed(2)} · Min step ${lot.min_step.toFixed(2)}</div>
        </div>

        ${lot.status === 'active' ? activeBlock(lot) : ''}
        ${lot.status === 'closed' ? closedBlock(lot) : ''}
      </div>

      <aside class="bg-white rounded-lg border border-slate-200 p-6">
        <h2 class="font-semibold mb-3">Place a bid</h2>
        <div id="bid-panel">${bidPanel(lot, user, flash)}</div>

        <h2 class="font-semibold mt-6 mb-3">Recent bids</h2>
        <ul id="bid-history" class="text-sm divide-y divide-slate-100">${bidHistory(bids)}</ul>

        ${
          !user
            ? `<div class="mt-6 text-xs text-slate-400 border-t border-slate-200 pt-3">
                Demo tip — sign up with any email/password to bid. Bots will keep bidding against you.
              </div>`
            : ''
        }
      </aside>
    </div>
  `;

  render(lot.title, body);
  wireUp(lotID);
}

function wireUp(lotID: string): void {
  const form = document.getElementById('bid-form') as HTMLFormElement | null;
  if (form) {
    form.addEventListener('submit', (e) => {
      e.preventDefault();
      const amt = parseFloat((document.getElementById('bid-amount') as HTMLInputElement).value);
      const user = currentUser();
      if (!user) return;
      try {
        placeBid(lotID, user.id, amt);
        const panel = document.getElementById('bid-panel');
        const lot = store.getLot(lotID);
        if (panel && lot) panel.innerHTML = bidPanel(lot, user, '');
      } catch (err) {
        const panel = document.getElementById('bid-panel');
        const lot = store.getLot(lotID);
        if (panel && lot) {
          panel.innerHTML = bidPanel(lot, user, err instanceof Error ? err.message : 'Bid failed');
        }
      }
    });
  }

  const tick = () => {
    const el = document.getElementById('time-remaining');
    if (!el) return;
    const closesAt = new Date(el.dataset.closesAt ?? '').getTime();
    el.textContent = formatRemaining(closesAt - Date.now());
  };
  tick();
  timerHandle = setInterval(tick, 1000);

  cleanup = onLot((ev) => {
    if (ev.type === 'new_bid' && ev.lot_id === lotID) {
      const cp = document.getElementById('current-price');
      if (cp) {
        cp.textContent = ev.new_price.toFixed(2);
        cp.classList.add('text-emerald-600');
        setTimeout(() => cp.classList.remove('text-emerald-600'), 700);
      }
      const list = document.getElementById('bid-history');
      if (list) {
        const empty = list.querySelector('.text-slate-400');
        if (empty) empty.remove();
        list.insertAdjacentHTML('afterbegin', bidRow(ev.bid));
        const first = list.firstElementChild;
        if (first) {
          first.classList.add('bg-amber-50', '-mx-1', 'px-1', 'rounded');
          setTimeout(() => first.classList.remove('bg-amber-50'), 1500);
        }
      }
      const lot = store.getLot(lotID);
      if (lot && lot.status === 'active') {
        const minBid = (lot.current_price + lot.min_step).toFixed(2);
        const input = document.getElementById('bid-amount') as HTMLInputElement | null;
        if (input && document.activeElement !== input) {
          input.value = minBid;
          input.min = minBid;
        }
      }
    } else if (ev.type === 'lot_extended' && ev.lot_id === lotID) {
      const el = document.getElementById('time-remaining');
      if (el) el.dataset.closesAt = ev.closing_at;
      const badge = document.getElementById('extended-badge');
      const cnt = document.getElementById('extended-count');
      if (badge && cnt) {
        badge.classList.remove('hidden');
        cnt.textContent = String(ev.extended_count);
        badge.classList.add('animate-pulse');
        setTimeout(() => badge.classList.remove('animate-pulse'), 1500);
      }
    } else if (ev.type === 'lot_closed' && ev.lot_id === lotID) {
      setTimeout(() => renderLot(lotID), 600);
    }
  });
}

export function lotPage(params: Record<string, string>): void {
  renderLot(params.id);
}

export function teardownLotPage(): void {
  teardown();
}
