<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { ArrowLeft, Globe, Network, Settings2, Shuffle, Sparkles } from "@lucide/svelte";
  import type { AppController } from "$lib/app/create-app.svelte";
  import CommunityInterfaces from "$lib/components/CommunityInterfaces.svelte";
  import ReticulumConfigEditor from "$lib/components/ReticulumConfigEditor.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();

  const suggestedCount = $derived(app.suggestedItems.length);
</script>

{#if app.initialSetupOpen}
  <div class="backdrop" aria-hidden="true"></div>
  <div
    class="dialog"
    role="dialog"
    aria-modal="true"
    aria-labelledby="initial-setup-title"
    aria-describedby="initial-setup-description"
  >
    <header class="header">
      <h2 id="initial-setup-title">{t("setup.title")}</h2>
      <p id="initial-setup-description" class="subtitle">{t("setup.subtitle")}</p>
    </header>

    {#if app.initialSetupError}
      <p class="error">{app.initialSetupError}</p>
    {/if}

    {#if app.initialSetupStep === "welcome"}
      <div class="choices">
        <button
          type="button"
          class="choice"
          disabled={app.initialSetupBusy}
          onclick={() => app.setInitialSetupStep("suggested")}
        >
          <span class="choice-icon suggested"><Shuffle size={18} /></span>
          <span class="choice-body">
            <span class="choice-title">{t("setup.suggestedTitle")}</span>
            <span class="choice-hint">{t("setup.suggestedHint", { count: 4 })}</span>
          </span>
        </button>
        <button
          type="button"
          class="choice"
          disabled={app.initialSetupBusy}
          onclick={() => app.setInitialSetupStep("pick")}
        >
          <span class="choice-icon pick"><Globe size={18} /></span>
          <span class="choice-body">
            <span class="choice-title">{t("setup.pickTitle")}</span>
            <span class="choice-hint">{t("setup.pickHint")}</span>
          </span>
        </button>
        <button
          type="button"
          class="choice"
          disabled={app.initialSetupBusy}
          onclick={() => app.setInitialSetupStep("config")}
        >
          <span class="choice-icon config"><Settings2 size={18} /></span>
          <span class="choice-body">
            <span class="choice-title">{t("setup.configTitle")}</span>
            <span class="choice-hint">{t("setup.configHint")}</span>
          </span>
        </button>
        <button
          type="button"
          class="choice subtle"
          disabled={app.initialSetupBusy}
          onclick={() => void app.skipInitialSetupAutoOnly()}
        >
          <span class="choice-icon auto"><Network size={18} /></span>
          <span class="choice-body">
            <span class="choice-title">{t("setup.autoOnlyTitle")}</span>
            <span class="choice-hint">{t("setup.autoOnlyHint")}</span>
          </span>
        </button>
      </div>
    {:else if app.initialSetupStep === "suggested"}
      <div class="step">
        <p class="step-hint">{t("setup.suggestedHint", { count: suggestedCount || 4 })}</p>
        <ul class="suggested-list">
          {#if app.suggestedLoading}
            <li class="empty">{t("community.loading")}</li>
          {:else if suggestedCount === 0}
            <li class="empty">{t("setup.suggestedEmpty")}</li>
          {:else}
            {#each app.suggestedItems as item (item.id)}
              <li>
                <span class="name">{item.name}</span>
                <span class="meta">{item.typeName} · {item.network}</span>
              </li>
            {/each}
          {/if}
        </ul>
        <div class="actions">
          <button
            type="button"
            class="ghost"
            disabled={app.initialSetupBusy}
            onclick={() => app.setInitialSetupStep("welcome")}
          >
            <ArrowLeft size={16} />
            {t("setup.back")}
          </button>
          <button
            type="button"
            class="ghost"
            disabled={app.initialSetupBusy || app.suggestedLoading}
            onclick={() => void app.loadSuggestedPreview()}
          >
            <Shuffle size={16} />
            {t("setup.refreshSuggested")}
          </button>
          <button
            type="button"
            class="primary"
            disabled={app.initialSetupBusy || app.suggestedLoading || suggestedCount === 0}
            onclick={() => void app.applySuggestedSetup()}
          >
            <Sparkles size={16} />
            {app.initialSetupBusy ? t("setup.busy") : t("setup.applySuggested")}
          </button>
        </div>
      </div>
    {:else if app.initialSetupStep === "pick"}
      <div class="step pick-step">
        <p class="step-hint">{t("setup.pickHint")}</p>
        <CommunityInterfaces
          items={app.communityItems}
          loading={app.communityLoading}
          importing={app.initialSetupBusy}
          error={app.communityError}
          fromBundle={app.communityFromBundle}
          bind:filter={app.communityFilter}
          selected={app.communitySelected}
          onFilter={(value) => {
            app.communityFilter = value;
          }}
          onRefresh={() => void app.loadCommunityInterfaces()}
          onToggle={app.toggleCommunitySelection}
          onImport={() => void app.importInitialSetupSelection()}
          showTitle={false}
        />
        <div class="actions">
          <button
            type="button"
            class="ghost"
            disabled={app.initialSetupBusy}
            onclick={() => app.setInitialSetupStep("welcome")}
          >
            <ArrowLeft size={16} />
            {t("setup.back")}
          </button>
        </div>
      </div>
    {:else}
      <div class="step config-step">
        <p class="step-hint">{t("setup.configHint")}</p>
        <ReticulumConfigEditor
          bind:configText={app.configText}
          configPath={app.configPath}
          saving={app.configSaving || app.initialSetupBusy}
          error={app.configError}
          showTitle={false}
          onChange={(text) => {
            app.configText = text;
          }}
          onSave={() => void app.saveInitialSetupConfig()}
          onReload={() => void app.reloadConfigFromDisk()}
          onOpenConfigDir={() => void app.openConfigFolder()}
        />
        <div class="actions">
          <button
            type="button"
            class="ghost"
            disabled={app.initialSetupBusy}
            onclick={() => app.setInitialSetupStep("welcome")}
          >
            <ArrowLeft size={16} />
            {t("setup.back")}
          </button>
          <button
            type="button"
            class="primary"
            disabled={app.initialSetupBusy || app.configSaving}
            onclick={() => void app.saveInitialSetupConfig()}
          >
            {app.initialSetupBusy ? t("setup.busy") : t("setup.saveAndContinue")}
          </button>
        </div>
      </div>
    {/if}
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 1300;
    background: rgb(0 0 0 / 0.55);
    cursor: default;
  }

  .dialog {
    position: fixed;
    top: 50%;
    left: 50%;
    z-index: 1301;
    width: min(42rem, calc(100vw - 1.5rem));
    max-height: min(88vh, 52rem);
    transform: translate(-50%, -50%);
    display: grid;
    gap: 0.9rem;
    padding: 1.15rem 1.2rem 1.1rem;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 4px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    overflow: auto;
  }

  .header h2 {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 650;
    color: var(--ren-fg);
  }

  .subtitle,
  .step-hint {
    margin: 0.35rem 0 0;
    font-size: 0.92rem;
    line-height: 1.45;
    color: var(--ren-fg-secondary);
  }

  .error {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.88rem;
  }

  .choices {
    display: grid;
    gap: 0.55rem;
  }

  .choice {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.75rem;
    align-items: start;
    width: 100%;
    padding: 0.8rem 0.85rem;
    border: 1px solid var(--ren-border);
    border-radius: 12px;
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    text-align: left;
    font: inherit;
    cursor: pointer;
    transition:
      border-color 0.15s ease,
      background 0.15s ease;
  }

  .choice:hover:not(:disabled) {
    border-color: color-mix(in srgb, var(--ren-accent) 45%, var(--ren-border));
    background: var(--ren-tab-hover);
  }

  .choice:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .choice.subtle {
    background: transparent;
  }

  .choice-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.1rem;
    height: 2.1rem;
    border-radius: 10px;
    color: var(--ren-fg);
    background: color-mix(in srgb, var(--ren-accent) 14%, transparent);
  }

  .choice-icon.auto {
    background: color-mix(in srgb, var(--ren-muted) 18%, transparent);
  }

  .choice-body {
    display: grid;
    gap: 0.2rem;
    min-width: 0;
  }

  .choice-title {
    font-size: 0.95rem;
    font-weight: 600;
  }

  .choice-hint {
    font-size: 0.84rem;
    line-height: 1.4;
    color: var(--ren-fg-secondary);
  }

  .step {
    display: grid;
    gap: 0.75rem;
    min-width: 0;
  }

  .pick-step,
  .config-step {
    min-width: 0;
  }

  .suggested-list {
    margin: 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 0.45rem;
    border: 1px solid var(--ren-border);
    border-radius: 12px;
    padding: 0.55rem;
    max-height: 14rem;
    overflow: auto;
  }

  .suggested-list li {
    display: grid;
    gap: 0.1rem;
    padding: 0.55rem 0.6rem;
    border-radius: 10px;
    background: var(--ren-input-bg);
  }

  .suggested-list .empty {
    background: transparent;
    color: var(--ren-fg-secondary);
    font-size: 0.9rem;
    text-align: center;
    padding: 1rem 0.5rem;
  }

  .name {
    font-size: 0.92rem;
    font-weight: 600;
    color: var(--ren-fg);
  }

  .meta {
    font-size: 0.8rem;
    color: var(--ren-muted);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.5rem;
    padding-top: 0.15rem;
  }

  .actions button {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    border: 1px solid var(--ren-border);
    border-radius: 10px;
    padding: 0.52rem 0.85rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    background: var(--ren-input-bg);
    color: var(--ren-fg);
  }

  .actions button:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .actions button.ghost:hover:not(:disabled) {
    background: var(--ren-tab-hover);
  }

  .actions button.primary {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
    color: #fff;
  }

  .actions button.primary:hover:not(:disabled) {
    background: var(--ren-accent-hover);
    border-color: var(--ren-accent-hover);
  }
</style>
