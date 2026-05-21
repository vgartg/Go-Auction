import { currentUser } from '../auth';
import { createLot } from '../engine';
import { render } from '../ui/layout';

function defaultClosingValue(): string {
  const d = new Date(Date.now() + 10 * 60 * 1000);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function form(errorMsg = ''): string {
  return `
    <div class="max-w-lg mx-auto bg-white border border-slate-200 rounded-lg p-6 mt-4">
      <h1 class="text-xl font-bold mb-4">Create a lot</h1>
      <form id="new-lot-form" class="space-y-3">
        <div>
          <label class="block text-xs text-slate-500 mb-1">Title</label>
          <input type="text" name="title" required maxlength="200"
            class="w-full border border-slate-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-slate-500 mb-1">Start price</label>
            <input type="number" name="start_price" step="0.01" min="0.01" required value="100.00"
              class="w-full border border-slate-300 rounded-md px-3 py-2 tabular-nums focus:outline-none focus:ring-2 focus:ring-amber-400"/>
          </div>
          <div>
            <label class="block text-xs text-slate-500 mb-1">Min step</label>
            <input type="number" name="min_step" step="0.01" min="0.01" required value="10.00"
              class="w-full border border-slate-300 rounded-md px-3 py-2 tabular-nums focus:outline-none focus:ring-2 focus:ring-amber-400"/>
          </div>
        </div>
        <div>
          <label class="block text-xs text-slate-500 mb-1">Closes at (your local time)</label>
          <input type="datetime-local" name="closing_at" required value="${defaultClosingValue()}"
            class="w-full border border-slate-300 rounded-md px-3 py-2 tabular-nums focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        ${errorMsg ? `<div class="text-sm text-red-600">${errorMsg}</div>` : ''}
        <button type="submit" class="w-full bg-amber-500 text-white rounded-md px-3 py-2 hover:bg-amber-600">Create</button>
      </form>
    </div>
  `;
}

export function newLotPage(): void {
  if (!currentUser()) {
    location.hash = '#/login?next=' + encodeURIComponent('/lots/new');
    return;
  }
  render('Create lot', form());
  attach('');
}

function attach(initialError: string): void {
  const f = document.getElementById('new-lot-form') as HTMLFormElement | null;
  if (!f) return;
  if (initialError) {
    render('Create lot', form(initialError));
    attach('');
    return;
  }
  f.addEventListener('submit', (e) => {
    e.preventDefault();
    const data = new FormData(f);
    const title = String(data.get('title') ?? '');
    const startPrice = parseFloat(String(data.get('start_price') ?? '0'));
    const minStep = parseFloat(String(data.get('min_step') ?? '0'));
    const closingRaw = String(data.get('closing_at') ?? '');
    const closingAt = new Date(closingRaw);
    if (isNaN(closingAt.getTime())) {
      render('Create lot', form('Invalid closing time'));
      attach('');
      return;
    }
    try {
      const lot = createLot({ title, start_price: startPrice, min_step: minStep, closing_at: closingAt });
      location.hash = `#/lots/${lot.id}`;
    } catch (err) {
      render('Create lot', form(err instanceof Error ? err.message : 'Failed to create'));
      attach('');
    }
  });
}
