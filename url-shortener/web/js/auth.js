/**
 * auth.js — shared controller for login.html and register.html.
 * Depends on api.js (window.Api).
 */
(function () {
  'use strict';

  const toast = document.getElementById('toast');
  function showToast(message, isError) {
    toast.textContent = message;
    toast.classList.toggle('is-error', !!isError);
    toast.classList.add('visible');
    clearTimeout(showToast._t);
    showToast._t = setTimeout(() => toast.classList.remove('visible'), 2600);
  }

  function nextDestination() {
    const params = new URLSearchParams(window.location.search);
    return params.get('next') === 'subscribe' ? 'index.html?subscribe=1' : 'index.html';
  }

  function setFieldError(el, message) {
    if (!message) {
      el.classList.add('hidden');
      el.textContent = '';
      return;
    }
    el.textContent = message;
    el.classList.remove('hidden');
  }

  // ---------- login form ----------
  const loginForm = document.getElementById('login-form');
  if (loginForm) {
    const submitBtn = document.getElementById('login-submit');
    const errorEl = document.getElementById('login-error');

    loginForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      setFieldError(errorEl, '');

      const email = document.getElementById('login-email').value.trim();
      const password = document.getElementById('login-password').value;

      if (!email || !password) {
        setFieldError(errorEl, 'Enter your email and password.');
        return;
      }

      submitBtn.disabled = true;
      submitBtn.textContent = 'Logging in…';
      try {
        await Api.login(email, password);
        window.location.href = nextDestination();
      } catch (err) {
        setFieldError(errorEl, err.status === 401 ? 'Incorrect email or password.' : err.message);
      } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Log in';
      }
    });
  }

  // ---------- register form ----------
  const registerForm = document.getElementById('register-form');
  if (registerForm) {
    const submitBtn = document.getElementById('register-submit');
    const errorEl = document.getElementById('register-error');

    registerForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      setFieldError(errorEl, '');

      const name = document.getElementById('reg-name').value.trim();
      const email = document.getElementById('reg-email').value.trim();
      const password = document.getElementById('reg-password').value;
      const confirm = document.getElementById('reg-password-confirm').value;
      const termsAccepted = document.getElementById('reg-terms').checked;

      if (!name || !email || !password) {
        setFieldError(errorEl, 'Fill in every field to continue.');
        return;
      }
      if (password.length < 8) {
        setFieldError(errorEl, 'Password must be at least 8 characters.');
        return;
      }
      if (password !== confirm) {
        setFieldError(errorEl, 'Passwords do not match.');
        return;
      }
      if (!termsAccepted) {
        setFieldError(errorEl, 'You need to accept the Terms to continue.');
        return;
      }

      submitBtn.disabled = true;
      submitBtn.textContent = 'Creating account…';
      try {
        await Api.register(name, email, password);
        window.location.href = nextDestination();
      } catch (err) {
        setFieldError(errorEl, err.status === 409 ? 'An account with that email already exists.' : err.message);
      } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Create account';
      }
    });
  }

  // ---------- OAuth buttons (shared) ----------
  // These do a real full-page redirect into the backend's OAuth
  // handler. The backend redirects to Google/LinkedIn's own login
  // and account-picker screen, then handles the callback and
  // redirects back here with a session cookie set — the picker UI
  // itself is rendered by Google/LinkedIn, not by this app.
  const redirectState = document.getElementById('redirect-state');
  const redirectText = document.getElementById('redirect-text');
  const mainForm = document.querySelector('#login-form, #register-form');

  function goToProvider(provider, label) {
    if (mainForm) mainForm.classList.add('hidden');
    if (redirectState) {
      redirectText.textContent = `Redirecting to ${label}…`;
      redirectState.classList.remove('hidden');
    }
    Api.startOAuth(provider);
  }

  const googleBtn = document.getElementById('oauth-google');
  const linkedinBtn = document.getElementById('oauth-linkedin');
  if (googleBtn) googleBtn.addEventListener('click', () => goToProvider('google', 'Google'));
  if (linkedinBtn) linkedinBtn.addEventListener('click', () => goToProvider('linkedin', 'LinkedIn'));
})();