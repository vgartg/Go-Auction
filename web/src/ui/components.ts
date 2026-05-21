import type { Bid, Lot, LotStatus } from '../types';
import { store } from '../store';
import { escape } from './layout';

export function statusBadge(s: LotStatus): string {
  if (s === 'active') {
    return `<span class="text-xs font-medium bg-emerald-100 text-emerald-700 rounded-full px-2 py-0.5">active</span>`;
  }
  if (s === 'closed') {
    return `<span class="text-xs font-medium bg-slate-200 text-slate-700 rounded-full px-2 py-0.5">closed</span>`;
  }
  return `<span class="text-xs font-medium bg-red-100 text-red-700 rounded-full px-2 py-0.5">${escape(s)}</span>`;
}

export function lotCard(lot: Lot): string {
  const closes = new Date(lot.closing_at);
  const closeLabel = lot.status === 'active' ? 'Closes' : 'Closed';
  return `
    <a href="#/lots/${lot.id}" class="block bg-white rounded-lg border border-slate-200 p-4 hover-lift">
      <div class="flex items-start justify-between gap-2 mb-3">
        <h3 class="font-semibold text-base line-clamp-2">${escape(lot.title)}</h3>
        ${statusBadge(lot.status)}
      </div>
      <div class="text-2xl font-bold text-slate-900 tabular-nums">${lot.current_price.toFixed(2)}</div>
      <div class="text-xs text-slate-500 mt-1">${closeLabel} ${escape(closes.toLocaleString())}</div>
      ${
        lot.extended_count > 0
          ? `<div class="mt-2 inline-block bg-amber-50 text-amber-700 text-xs font-medium rounded px-2 py-0.5">Extended × ${lot.extended_count}</div>`
          : ''
      }
    </a>
  `;
}

export function filterPill(label: string, href: string, active: boolean): string {
  return active
    ? `<a href="${href}" class="px-3 py-1 rounded-full bg-slate-900 text-white">${escape(label)}</a>`
    : `<a href="${href}" class="px-3 py-1 rounded-full bg-white border border-slate-200 text-slate-700 hover:border-slate-400">${escape(label)}</a>`;
}

export function bidderLabel(userID: string): string {
  const u = store.getUser(userID);
  return u ? u.username : `${userID.slice(0, 8)}…`;
}

export function bidRow(b: Bid): string {
  const t = new Date(b.created_at);
  const time = `${t.getHours()}:${String(t.getMinutes()).padStart(2, '0')}`;
  return `
    <li class="py-2 flex justify-between gap-2">
      <a href="#/users/${b.user_id}" class="text-slate-500 hover:text-slate-900 text-xs">
        ${escape(bidderLabel(b.user_id))}
      </a>
      <div class="text-right">
        <div class="font-semibold tabular-nums">${b.amount.toFixed(2)}</div>
        <div class="text-xs text-slate-400">${time}</div>
      </div>
    </li>
  `;
}
