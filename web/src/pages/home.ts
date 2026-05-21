import { currentUser } from '../auth';
import { store } from '../store';
import { filterPill, lotCard } from '../ui/components';
import { render } from '../ui/layout';

export function homePage(_params: Record<string, string>, query: URLSearchParams): void {
  const status = query.get('status') ?? '';
  const lots = store.listLots(status || undefined);
  const user = currentUser();

  const filterRow = `
    <div class="flex items-center gap-2 mb-6 text-sm">
      ${filterPill('All', '#/', status === '')}
      ${filterPill('Active', '#/?status=active', status === 'active')}
      ${filterPill('Closed', '#/?status=closed', status === 'closed')}
    </div>
  `;

  const header = `
    <div class="flex items-center justify-between mb-4">
      <div>
        <h1 class="text-2xl font-bold">Auctions</h1>
        <p class="text-sm text-slate-500">Live bidding. Optimistic locking. Anti-sniping.</p>
      </div>
      ${
        user
          ? `<a href="#/lots/new" class="bg-amber-500 hover:bg-amber-600 text-white rounded-md px-4 py-2 text-sm font-medium">+ Create lot</a>`
          : ''
      }
    </div>
  `;

  const grid =
    lots.length === 0
      ? `<div class="rounded-lg border border-dashed border-slate-300 bg-white p-12 text-center text-slate-500">
          No lots yet.
          ${
            user
              ? `<a href="#/lots/new" class="text-amber-600 hover:underline">Create the first one →</a>`
              : `<a href="#/register" class="text-amber-600 hover:underline">Sign up to create one →</a>`
          }
        </div>`
      : `<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">${lots.map(lotCard).join('')}</div>`;

  render('Auctions', header + filterRow + grid);
}
