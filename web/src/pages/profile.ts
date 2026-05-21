import { store } from '../store';
import { escape, render } from '../ui/layout';

function initials(name: string): string {
  return name ? name[0].toUpperCase() : '?';
}

function tile(label: string, value: string): string {
  return `
    <div class="bg-slate-50 rounded-md p-4">
      <div class="text-xs uppercase tracking-wide text-slate-500">${escape(label)}</div>
      <div class="text-2xl font-bold tabular-nums mt-1">${escape(value)}</div>
    </div>
  `;
}

export function profilePage(params: Record<string, string>): void {
  const user = store.getUser(params.id);
  if (!user) {
    render('Not found', `<div class="text-center text-slate-500 py-12">User not found.</div>`);
    return;
  }
  const bids = store.bidsByUser(user.id);
  const wins = store.winsByUser(user.id);
  const totalSpent = wins.reduce((sum, l) => sum + l.current_price, 0);

  const body = `
    <div class="bg-white border border-slate-200 rounded-lg p-6 max-w-2xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-12 h-12 rounded-full bg-gradient-to-br from-amber-400 to-amber-600 flex items-center justify-center text-white font-bold text-lg">
          ${escape(initials(user.username))}
        </div>
        <div>
          <h1 class="text-xl font-bold">${escape(user.username)}</h1>
          <div class="text-xs text-slate-500 font-mono">${escape(user.id)}</div>
        </div>
      </div>
      <div class="grid grid-cols-3 gap-3">
        ${tile('Bids placed', String(bids.length))}
        ${tile('Lots won', String(wins.length))}
        ${tile('Total spent', totalSpent.toFixed(2))}
      </div>
    </div>
  `;
  render(user.username, body);
}
