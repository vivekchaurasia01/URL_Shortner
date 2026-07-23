/**
 * api.js
 * Thin fetch wrapper around the short.do Go backend.
 *
 * Expected backend contract (adjust paths to match your router):
 *
 *   POST   /api/shorten          { url, alias? }        -> { code, short_url, expires_at }
 *   GET    /api/me                                       -> { id, name, email, plan } | 401
 *   POST   /api/auth/login       { email, password }     -> { token, user } , sets httpOnly cookie
 *   POST   /api/auth/register    { name, email, password }
 *   POST   /api/auth/logout
 *   GET    /auth/google/login                             -> 302 redirect to Google, then back to /auth/callback
 *   GET    /auth/linkedin/login                            -> 302 redirect to LinkedIn, then back to /auth/callback
 *   POST   /api/billing/subscribe                          -> { checkout_url }  (e.g. Stripe Checkout session)
 *
 * Auth model assumed: backend sets an httpOnly session cookie on
 * login/OAuth callback, so the browser never has to hold a token.
 * `credentials: 'include'` is set on every request for that reason.
 * If you instead return a bearer token, keep it in memory only
 * (not localStorage) and attach it in an Authorization header here.
 */

const API_BASE = window.SHORTDO_API_BASE || ''; // e.g. 'https://api.short.do' — '' uses same origin

class ApiError extends Error {
  constructor(message, status, payload) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.payload = payload;
  }
}

async function request(path, { method = 'GET', body, headers = {} } = {}) {
  let res;
  try {
    res = await fetch(`${API_BASE}${path}`, {
      method,
      credentials: 'include',
      headers: {
        ...(body ? { 'Content-Type': 'application/json' } : {}),
        Accept: 'application/json',
        ...headers,
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch (networkErr) {
    throw new ApiError('Could not reach the server. Check your connection.', 0, null);
  }

  const isJson = res.headers.get('content-type')?.includes('application/json');
  const data = isJson ? await res.json().catch(() => null) : null;

  if (!res.ok) {
    const message = data?.message || data?.error || `Request failed (${res.status})`;
    throw new ApiError(message, res.status, data);
  }
  return data;
}

const Api = {
  shorten(url, alias) {
    return request('/api/shorten', { method: 'POST', body: alias ? { url, alias } : { url } });
  },

  me() {
    return request('/api/me').catch((err) => {
      if (err.status === 401) return null;
      throw err;
    });
  },

  login(email, password) {
    return request('/api/auth/login', { method: 'POST', body: { email, password } });
  },

  register(name, email, password) {
    return request('/api/auth/register', { method: 'POST', body: { name, email, password } });
  },

  logout() {
    return request('/api/auth/logout', { method: 'POST' });
  },

  subscribe() {
    return request('/api/billing/subscribe', { method: 'POST' });
  },

  /** Full-page redirect into the backend's OAuth flow. The backend
   *  itself redirects to Google/LinkedIn, then handles the callback
   *  and redirects back to the app with a session cookie set. */
  startOAuth(provider) {
    window.location.href = `${API_BASE}/auth/${provider}/login`;
  },
};

window.Api = Api;
window.ApiError = ApiError;