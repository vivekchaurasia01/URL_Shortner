/**
 * app.js — index page controller.
 * Depends on api.js (window.Api) being loaded first.
 */
(function () {
  'use strict';

  const form = document.getElementById('shorten-form');
  const urlInput = document.getElementById('url-input');
  const urlError = document.getElementById('url-error');
  const shortenBtn = document.getElementById('shorten-btn');

  const aliasToggle = document.getElementById('alias-toggle');
  const aliasRow = document.getElementById('alias-row');
  const aliasInput = document.getElementById('alias-input');
  const aliasProBadge = document.getElementById('alias-pro-badge');

  const resultCard = document.getElementById('result-card');
  const resultLink = document.getElementById('result-link');
  const resultMeta = document.getElementById('result-meta');
  const copyBtn = document.getElementById('copy-btn');
  const qrBtn = document.getElementById('qr-btn');
  const qrPanel = document.getElementById('qr-panel');
  const qrCanvas = document.getElementById('qr-canvas');

  const headerActions = document.getElementById('header-actions');
  const freeLoginPerk = document.getElementById('free-login-perk');
  const freeCta = document.getElementById('free-cta');
  const proCta = document.getElementById('pro-cta');
  const trustLine = document.getElementById('trust-line');
  const toast = document.getElementById('toast');

  document.getElementById('year').textContent = new Date().getFullYear();

  // Falls back to localhost:8081 if index.html doesn't define it —
  // set window.SHORTDO_REDIRECTOR_BASE in index.html to override.
  const REDIRECTOR_BASE = window.SHORTDO_REDIRECTOR_BASE || 'http://127.0.0.1:8081';

  let session = null; // { id, name, email, plan: 'free' | 'pro' } | null
  let statsPollTimer = null;

  function showToast(message, isError) {
    toast.textContent = message;
    toast.classList.toggle('is-error', !!isError);
    toast.classList.add('visible');
    clearTimeout(showToast._t);
    showToast._t = setTimeout(() => toast.classList.remove('visible'), 2600);
  }

  function isValidUrl(value) {
    try {
      const u = new URL(value);
      return u.protocol === 'http:' || u.protocol === 'https:';
    } catch {
      return false;
    }
  }

  // ---------- session-dependent UI ----------
  function renderForSession() {
    const isPro = session?.plan === 'pro';
    const isLoggedIn = !!session;

    headerActions.innerHTML = isLoggedIn
      ? `<span>${session.name}</span> <button class="btn btn-ghost btn-sm" id="logout-btn">Log out</button>`
      : `<a href="login.html">Log in</a>`;
    if (isLoggedIn) {
      document.getElementById('logout-btn').addEventListener('click', async () => {
        try {
          await Api.logout();
          session = null;
          renderForSession();
          showToast('Logged out');
        } catch (err) {
          showToast(err.message, true);
        }
      });
    }

    aliasToggle.disabled = !isPro;
    aliasProBadge.classList.toggle('hidden', isPro);
    if (!isPro) aliasRow.classList.remove('visible');

    freeLoginPerk.classList.toggle('hidden', isLoggedIn);
    freeCta.textContent = isLoggedIn ? 'Logged in' : 'Log in';
    freeCta.disabled = isLoggedIn;

    proCta.textContent = isPro ? 'Current plan' : 'Subscribe';
    proCta.disabled = isPro;

    trustLine.textContent = isLoggedIn
      ? isPro
        ? 'Unlimited links · never expire · QR, analytics and custom URLs unlocked'
        : '3 free links a week as a logged-in user · Pro links come with QR, analytics and custom URLs'
      : '2 free links a week (3 if logged in) · Pro links come with QR, analytics and custom URLs';
  }

  freeCta.addEventListener('click', () => {
    window.location.href = 'login.html';
  });

  proCta.addEventListener('click', async () => {
    if (!session) {
      window.location.href = 'login.html?next=subscribe';
      return;
    }
    proCta.disabled = true;
    proCta.textContent = 'Redirecting…';
    try {
      const { checkout_url } = await Api.subscribe();
      window.location.href = checkout_url;
    } catch (err) {
      proCta.disabled = false;
      proCta.textContent = 'Subscribe';
      showToast(err.message, true);
    }
  });

  aliasToggle.addEventListener('click', () => {
    if (aliasToggle.disabled) {
      showToast('Custom aliases are a Pro feature');
      return;
    }
    aliasRow.classList.toggle('visible');
    if (aliasRow.classList.contains('visible')) aliasInput.focus();
  });

  // ---------- shorten flow ----------
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const url = urlInput.value.trim();

    if (!isValidUrl(url)) {
      urlInput.classList.add('has-error');
      urlError.classList.add('visible');
      urlInput.focus();
      return;
    }
    urlInput.classList.remove('has-error');
    urlError.classList.remove('visible');

    const alias = aliasRow.classList.contains('visible') ? aliasInput.value.trim() : undefined;

    shortenBtn.disabled = true;
    shortenBtn.textContent = 'Shortening…';
    try {
      const data = await Api.shorten(url, alias || undefined);
      showResult(data);
    } catch (err) {
      showToast(err.message, true);
    } finally {
      shortenBtn.disabled = false;
      shortenBtn.textContent = 'Shorten';
    }
  });

function showResult(data) {
  // Stop any polling from a previous result before starting a new one.
  if (statsPollTimer) {
    clearInterval(statsPollTimer);
    statsPollTimer = null;
  }

  const code = data.code || data.short_url;
  const displayUrl = `${REDIRECTOR_BASE}/${code}`;

  resultLink.textContent = code;
  resultLink.href = displayUrl;

  const isPro = session?.plan === 'pro';
  const expiryText = isPro ? 'never expires' : 'expires in 24h if unused';

  function refreshStats() {
    Api.getStats(code)
      .then((stats) => {
        resultMeta.textContent = `${stats.clicks} click${stats.clicks === 1 ? '' : 's'} · ${expiryText}`;
      })
      .catch(() => {
        resultMeta.textContent = `0 clicks · ${expiryText}`;
      });
  }

  resultMeta.textContent = `loading clicks… · ${expiryText}`;
  refreshStats();

  // Poll every 5s so the count updates live while this result is visible.
  statsPollTimer = setInterval(refreshStats, 5000);

  qrPanel.classList.remove('visible');
  resultCard.classList.add('visible');
  resultCard.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

  copyBtn.addEventListener('click', async () => {
    try {
      await navigator.clipboard.writeText(resultLink.href);
      showToast('Copied to clipboard');
    } catch {
      showToast('Could not copy — copy it manually', true);
    }
  });

  qrBtn.addEventListener('click', () => {
    if (session?.plan !== 'pro') {
      showToast('QR codes are a Pro feature');
      window.location.href = 'login.html?next=subscribe';
      return;
    }
    const isVisible = qrPanel.classList.toggle('visible');
    if (isVisible && window.QRCode) {
      QRCode.toCanvas(qrCanvas, resultLink.href, { width: 168, margin: 1 }, (err) => {
        if (err) showToast('Could not generate QR code', true);
      });
    }
  });

  // ---------- init ----------
  (async function init() {
    try {
      session = await Api.me();
    } catch {
      session = null;
    }
    renderForSession();
  })();
})();