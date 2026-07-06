<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Fingerprint, Home, Plus, Square } from "@lucide/svelte";
  import { MAX_TABS } from "$lib/browser/url";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    url: string;
    tabCount: number;
    canIdentify: boolean;
    identifying: boolean;
    atTabLimit: boolean;
    onNavigate: (url: string) => void;
    onHome: () => void;
    onNewTab: () => void;
    onOpenTabs: () => void;
    onIdentify: () => void;
  };

  let {
    url = $bindable(""),
    tabCount,
    canIdentify,
    identifying,
    atTabLimit,
    onNavigate,
    onHome,
    onNewTab,
    onOpenTabs,
    onIdentify,
  }: Props = $props();

  function submit(event: Event) {
    event.preventDefault();
    onNavigate(url);
  }
</script>

<header class="mobile-chrome">
  <button class="ren-icon-btn" type="button" aria-label={t("mobileNav.browse")} onclick={onHome}>
    <Home size={18} />
  </button>

  <form class="url-form" onsubmit={submit}>
    <input
      class="url-input ren-input"
      bind:value={url}
      placeholder={t("chrome.urlPlaceholder")}
      spellcheck="false"
      autocomplete="off"
    />
  </form>

  {#if canIdentify}
    <button
      class="ren-icon-btn identify-btn"
      type="button"
      aria-label={t("chrome.identify")}
      title={t("chrome.identifyTitle")}
      disabled={identifying}
      onclick={onIdentify}
    >
      <Fingerprint size={16} />
    </button>
  {/if}

  <button
    class="ren-icon-btn"
    type="button"
    aria-label={t("tab.newTab")}
    disabled={atTabLimit}
    title={atTabLimit ? t("tab.limitReached", { max: MAX_TABS }) : t("tab.newTab")}
    onclick={onNewTab}
  >
    <Plus size={18} />
  </button>

  <button
    class="ren-icon-btn tabs-btn"
    type="button"
    aria-label={t("mobileTabs.switcher")}
    onclick={onOpenTabs}
  >
    <Square size={16} />
    <span class="tab-count" aria-hidden="true">{tabCount}</span>
  </button>
</header>

<style>
  .mobile-chrome {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.5rem 0.65rem;
    padding-top: calc(0.5rem + env(safe-area-inset-top));
    background: var(--ren-chrome-bg);
    border-bottom: 1px solid var(--ren-border);
  }

  .url-form {
    flex: 1;
    min-width: 0;
  }

  .url-input {
    width: 100%;
    border-radius: 999px;
    background: var(--ren-surface-muted);
    border-color: var(--ren-border);
    font-size: 0.88rem;
    padding: 0.45rem 0.75rem;
  }

  .identify-btn {
    flex-shrink: 0;
  }

  .tabs-btn {
    position: relative;
    flex-shrink: 0;
  }

  .tab-count {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    font-size: 0.62rem;
    font-weight: 700;
    line-height: 1;
    pointer-events: none;
  }
</style>
