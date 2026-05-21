import { currentUser, logout } from '../auth';

export function escape(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

export function shell(title: string, body: string, flash?: string): string {
  document.title = `${title} · GoAuction`;
  const user = currentUser();
  return `
    <header class="bg-white border-b border-slate-200 sticky top-0 z-30">
      <div class="max-w-6xl mx-auto px-4 h-14 flex items-center justify-between">
        <a href="#/" class="flex items-center gap-2 font-bold text-lg">
          <span>GoAuction</span>
          <span class="hidden sm:inline-block text-[10px] uppercase tracking-wider bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">demo</span>
        </a>
        <nav class="flex items-center gap-4 text-sm">
          <a href="#/" class="text-slate-700 hover:text-slate-900">Lots</a>
          ${
            user
              ? `
            <a href="#/lots/new" class="text-slate-700 hover:text-slate-900">Create lot</a>
            <a href="#/users/${user.id}" class="text-slate-700 hover:text-slate-900">${escape(user.username)}</a>
            <button id="logout-btn" class="text-slate-500 hover:text-slate-900">Log out</button>
          `
              : `
            <a href="#/login" class="text-slate-700 hover:text-slate-900">Log in</a>
            <a href="#/register" class="bg-slate-900 text-white rounded-md px-3 py-1.5 hover:bg-slate-800">Sign up</a>
          `
          }
        </nav>
      </div>
    </header>
    <main class="flex-1 w-full max-w-6xl mx-auto px-4 py-6 fade-in">
      ${
        flash
          ? `<div class="mb-4 rounded-md bg-red-50 border border-red-200 text-red-800 px-4 py-2 text-sm">${escape(flash)}</div>`
          : ''
      }
      ${body}
    </main>
    <footer class="mt-10 py-6 text-center text-xs text-slate-400">
      <a href="https://github.com/vgartg/goauction" class="hover:text-slate-600">vgartg/goauction</a>
      <span class="mx-2">·</span>
      <span>Go + templ + HTMX + WebSocket — this page is a client-side demo</span>
    </footer>
  `;
}

export function render(title: string, body: string, flash?: string): void {
  const app = document.getElementById('app');
  if (!app) return;
  app.innerHTML = shell(title, body, flash);
  const lo = document.getElementById('logout-btn');
  if (lo) {
    lo.addEventListener('click', () => {
      logout();
      location.hash = '#/';
      location.reload();
    });
  }
}
