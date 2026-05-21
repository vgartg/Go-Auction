import { startBots } from './bots';
import { restoreTimers } from './engine';
import { onLot } from './events';
import { homePage } from './pages/home';
import { loginPage } from './pages/login';
import { lotPage, teardownLotPage } from './pages/lot';
import { newLotPage } from './pages/new_lot';
import { profilePage } from './pages/profile';
import { registerPage } from './pages/register';
import { route, startRouter } from './router';
import { seedIfEmpty } from './seed';

let currentPath = '';

function track(handler: (params: Record<string, string>, query: URLSearchParams) => void) {
  return (params: Record<string, string>, query: URLSearchParams) => {
    if (currentPath.startsWith('/lots/') && !currentPath.endsWith('/new')) {
      teardownLotPage();
    }
    const path = location.hash.slice(1).split('?')[0] || '/';
    currentPath = path;
    handler(params, query);
  };
}

function init(): void {
  seedIfEmpty();
  restoreTimers();

  route('/', track(homePage));
  route('/lots/new', track(newLotPage));
  route('/lots/:id', track(lotPage));
  route('/login', track(loginPage));
  route('/register', track(registerPage));
  route('/users/:id', track(profilePage));

  onLot((ev) => {
    if (ev.type === 'lot_changed' && currentPath === '/') {
      const q = new URLSearchParams(location.hash.split('?')[1] ?? '');
      homePage({}, q);
    }
  });

  startRouter();
  startBots();
}

init();
