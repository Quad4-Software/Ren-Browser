<script lang="ts">
  import { Moon, Sun } from "@lucide/svelte";
  import { displayName } from "$lib/brand";
  import { login } from "$lib/auth/api";
  import { applyTheme, defaultTheme, type ThemeSettings } from "$lib/theme/tokens";

  const THEME_KEY = "renbrowser:auth-theme";

  let password = $state("");
  let error = $state("");
  let blocked = $state(false);
  let retryIn = $state(0);
  let loading = $state(false);
  let theme = $state<ThemeSettings>(loadTheme());

  function loadTheme(): ThemeSettings {
    try {
      const raw = localStorage.getItem(THEME_KEY);
      if (raw) {
        return { ...defaultTheme(), ...JSON.parse(raw) };
      }
    } catch {
      /* ignore */
    }
    return defaultTheme();
  }

  function resolvedMode(current = theme): "dark" | "light" {
    if (current.mode === "system") {
      return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    }
    return current.mode;
  }

  $effect(() => {
    applyTheme(theme);
  });

  function toggleTheme() {
    const nextMode = resolvedMode() === "dark" ? "light" : "dark";
    theme = { ...theme, mode: nextMode };
    localStorage.setItem(THEME_KEY, JSON.stringify(theme));
  }

  async function submit(event: Event) {
    event.preventDefault();
    if (loading || blocked) {
      return;
    }
    loading = true;
    error = "";
    try {
      const result = await login(password);
      if (result.ok) {
        window.location.reload();
        return;
      }
      blocked = Boolean(result.blocked);
      retryIn = result.retryIn ?? 0;
      error = result.error ?? "login failed";
    } catch {
      error = "network error";
    } finally {
      loading = false;
    }
  }
</script>

<div class="auth-shell">
  <button
    type="button"
    class="theme-btn ren-icon-btn"
    aria-label="Toggle theme"
    onclick={toggleTheme}
  >
    {#if resolvedMode() === "dark"}
      <Sun size={18} />
    {:else}
      <Moon size={18} />
    {/if}
  </button>

  <main class="auth-card">
    <h1>{displayName}</h1>
    <p class="subtitle">Sign in to continue</p>

    <form onsubmit={submit}>
      <label class="field">
        <span>Password</span>
        <input
          type="password"
          autocomplete="current-password"
          bind:value={password}
          disabled={loading || blocked}
        />
      </label>

      {#if error}
        <p class="error" role="alert">
          {error}
          {#if blocked && retryIn > 0}
            <span class="retry">Try again in {retryIn}s.</span>
          {/if}
        </p>
      {/if}

      <button type="submit" class="submit" disabled={loading || blocked || !password}>
        {loading ? "Signing in..." : "Sign in"}
      </button>
    </form>
  </main>
</div>

<style>
  .auth-shell {
    min-height: 100dvh;
    display: grid;
    place-items: center;
    padding: 1.5rem;
    background: var(--ren-surface-bg);
    color: var(--ren-fg);
    position: relative;
  }

  .theme-btn {
    position: absolute;
    top: 1rem;
    right: 1rem;
  }

  .auth-card {
    width: min(100%, 24rem);
    padding: 1.75rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
  }

  h1 {
    margin: 0;
    font-size: 1.35rem;
    font-weight: 650;
  }

  .subtitle {
    margin: 0.35rem 0 1.25rem;
    color: var(--ren-muted);
    font-size: 0.95rem;
  }

  .field {
    display: grid;
    gap: 0.4rem;
    margin-bottom: 1rem;
    font-size: 0.9rem;
  }

  input {
    width: 100%;
    padding: 0.65rem 0.75rem;
    border-radius: 0.65rem;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
  }

  input:focus {
    outline: 2px solid var(--ren-focus);
    outline-offset: 1px;
  }

  .submit {
    width: 100%;
    padding: 0.7rem 1rem;
    border: none;
    border-radius: 0.65rem;
    background: var(--ren-accent);
    color: white;
    font-weight: 600;
    cursor: pointer;
  }

  .submit:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .error {
    margin: 0 0 0.85rem;
    color: var(--ren-danger);
    font-size: 0.9rem;
  }

  .retry {
    display: block;
    margin-top: 0.25rem;
    color: var(--ren-muted);
  }
</style>
