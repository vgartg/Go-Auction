type Handler = (params: Record<string, string>, query: URLSearchParams) => void;

interface Route {
  pattern: RegExp;
  keys: string[];
  handler: Handler;
}

const routes: Route[] = [];

function compile(path: string): { pattern: RegExp; keys: string[] } {
  const keys: string[] = [];
  const re = path
    .replace(/\//g, '\\/')
    .replace(/:([a-zA-Z_]+)/g, (_m, k: string) => {
      keys.push(k);
      return '([^/]+)';
    });
  return { pattern: new RegExp('^' + re + '$'), keys };
}

export function route(path: string, handler: Handler): void {
  const { pattern, keys } = compile(path);
  routes.push({ pattern, keys, handler });
}

export function go(hash: string): void {
  if (location.hash === hash) {
    dispatch();
  } else {
    location.hash = hash;
  }
}

function parseHash(): { path: string; query: URLSearchParams } {
  const raw = location.hash.startsWith('#') ? location.hash.slice(1) : '/';
  const [path, qs] = raw.split('?');
  return {
    path: path || '/',
    query: new URLSearchParams(qs ?? ''),
  };
}

function dispatch(): void {
  const { path, query } = parseHash();
  for (const r of routes) {
    const m = r.pattern.exec(path);
    if (m) {
      const params: Record<string, string> = {};
      r.keys.forEach((k, i) => (params[k] = decodeURIComponent(m[i + 1])));
      r.handler(params, query);
      return;
    }
  }
  routes[0]?.handler({}, query);
}

export function startRouter(): void {
  window.addEventListener('hashchange', dispatch);
  if (!location.hash) location.replace('#/');
  else dispatch();
}
