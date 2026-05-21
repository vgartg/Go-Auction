import { store } from './store';
import type { User } from './types';
import { uuid } from './uuid';

const SESSION_KEY = 'goauction:session:v1';

function hash(password: string): string {
  let h = 0x811c9dc5;
  for (let i = 0; i < password.length; i++) {
    h ^= password.charCodeAt(i);
    h = (h + ((h << 1) + (h << 4) + (h << 7) + (h << 8) + (h << 24))) >>> 0;
  }
  return `demo$${h.toString(16)}`;
}

export function currentUserID(): string | null {
  return localStorage.getItem(SESSION_KEY);
}

export function currentUser(): User | null {
  const id = currentUserID();
  return id ? store.getUser(id) : null;
}

export function register(username: string, email: string, password: string): User {
  const u = username.trim();
  const e = email.trim().toLowerCase();
  if (u.length < 2) throw new Error('username too short');
  if (!e.includes('@')) throw new Error('invalid email');
  if (password.length < 6) throw new Error('password too short');
  if (store.findUserByEmail(e) || store.findUserByUsername(u)) {
    throw new Error('Username or email already taken');
  }
  const user: User = {
    id: uuid(),
    username: u,
    email: e,
    password_hash: hash(password),
    created_at: new Date().toISOString(),
  };
  store.addUser(user);
  localStorage.setItem(SESSION_KEY, user.id);
  return user;
}

export function login(email: string, password: string): User {
  const user = store.findUserByEmail(email.trim().toLowerCase());
  if (!user || user.password_hash !== hash(password)) {
    throw new Error('Invalid credentials');
  }
  localStorage.setItem(SESSION_KEY, user.id);
  return user;
}

export function logout(): void {
  localStorage.removeItem(SESSION_KEY);
}

export function loginAs(userID: string): void {
  if (!store.getUser(userID)) return;
  localStorage.setItem(SESSION_KEY, userID);
}
