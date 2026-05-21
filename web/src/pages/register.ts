import { register } from '../auth';
import { escape, render } from '../ui/layout';

function form(errorMsg = '', username = '', email = ''): string {
  return `
    <div class="max-w-sm mx-auto bg-white border border-slate-200 rounded-lg p-6 mt-8">
      <h1 class="text-xl font-bold mb-4">Sign up</h1>
      <form id="register-form" class="space-y-3">
        <div>
          <label class="block text-xs text-slate-500 mb-1">Username</label>
          <input type="text" name="username" value="${escape(username)}" required minlength="2" maxlength="64"
            class="w-full border border-slate-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        <div>
          <label class="block text-xs text-slate-500 mb-1">Email</label>
          <input type="email" name="email" value="${escape(email)}" required
            class="w-full border border-slate-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        <div>
          <label class="block text-xs text-slate-500 mb-1">Password (min 6)</label>
          <input type="password" name="password" required minlength="6"
            class="w-full border border-slate-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400"/>
        </div>
        ${errorMsg ? `<div class="text-sm text-red-600">${escape(errorMsg)}</div>` : ''}
        <button type="submit" class="w-full bg-amber-500 text-white rounded-md px-3 py-2 hover:bg-amber-600">Create account</button>
      </form>
      <p class="text-sm text-slate-500 mt-4 text-center">
        Already a member? <a href="#/login" class="text-amber-600 hover:underline">Log in</a>
      </p>
    </div>
  `;
}

export function registerPage(): void {
  render('Sign up', form());
  attach();
}

function attach(): void {
  const f = document.getElementById('register-form') as HTMLFormElement | null;
  if (!f) return;
  f.addEventListener('submit', (e) => {
    e.preventDefault();
    const data = new FormData(f);
    const username = String(data.get('username') ?? '');
    const email = String(data.get('email') ?? '');
    try {
      register(username, email, String(data.get('password') ?? ''));
      location.hash = '#/';
      location.reload();
    } catch (err) {
      render('Sign up', form(err instanceof Error ? err.message : 'Failed', username, email));
      attach();
    }
  });
}
