import { login } from '../auth';
import { escape, render } from '../ui/layout';

function form(next: string, errorMsg = '', email = ''): string {
  return `
    <div class="max-w-sm mx-auto bg-white border border-slate-200 rounded-lg p-6 mt-8">
      <h1 class="text-xl font-bold mb-4">Log in</h1>
      <form id="login-form" class="space-y-3">
        <input type="hidden" name="next" value="${escape(next)}"/>
        <div>
          <label class="block text-xs text-slate-500 mb-1">Email</label>
          <input type="email" name="email" value="${escape(email)}" required
            class="w-full border border-slate-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        <div>
          <label class="block text-xs text-slate-500 mb-1">Password</label>
          <input type="password" name="password" required
            class="w-full border border-slate-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        ${errorMsg ? `<div class="text-sm text-red-600">${escape(errorMsg)}</div>` : ''}
        <button type="submit" class="w-full bg-slate-900 text-white rounded-md px-3 py-2 hover:bg-slate-800">Log in</button>
      </form>
      <p class="text-sm text-slate-500 mt-4 text-center">
        No account? <a href="#/register" class="text-amber-600 hover:underline">Sign up</a>
      </p>
    </div>
  `;
}

export function loginPage(_params: Record<string, string>, query: URLSearchParams): void {
  const next = query.get('next') ?? '/';
  render('Log in', form(next));
  attach(next);
}

function attach(next: string): void {
  const f = document.getElementById('login-form') as HTMLFormElement | null;
  if (!f) return;
  f.addEventListener('submit', (e) => {
    e.preventDefault();
    const data = new FormData(f);
    try {
      login(String(data.get('email') ?? ''), String(data.get('password') ?? ''));
      location.hash = `#${next.startsWith('/') ? next : '/' + next}`;
      location.reload();
    } catch (err) {
      render('Log in', form(next, err instanceof Error ? err.message : 'Login failed', String(data.get('email') ?? '')));
      attach(next);
    }
  });
}
