<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";
  import { GetNodeIdentifyOnConnect } from "../../../bindings/renbrowser/internal/app/browserservice.js";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();

  let identifyAlways = $state(false);
  let identifyLoadedFor = $state("");

  $effect(() => {
    if (!app.identifyConfirmOpen || identifyLoadedFor === app.url) {
      return;
    }
    const forURL = app.url;
    identifyLoadedFor = forURL;
    void GetNodeIdentifyOnConnect(forURL)
      .then((on) => {
        if (app.url === forURL) {
          identifyAlways = !!on;
        }
      })
      .catch(() => {});
  });
</script>

<ConfirmDialog
  open={app.identifyConfirmOpen}
  title={t("dialog.identifyTitle")}
  message={t("dialog.identifyMessage")}
  confirmLabel={t("common.identify")}
  onConfirm={() => app.confirmIdentify(identifyAlways)}
  onCancel={() => (app.identifyConfirmOpen = false)}
>
  <label class="identify-always">
    <input type="checkbox" bind:checked={identifyAlways} />
    <span>{t("dialog.identifyAlways")}</span>
  </label>
</ConfirmDialog>

<ConfirmDialog
  open={app.resetDbConfirmOpen}
  title={t("dialog.resetDbTitle")}
  message={t("dialog.resetDbMessage")}
  confirmLabel={t("dialog.resetDbConfirm")}
  onConfirm={app.confirmResetDatabase}
  onCancel={() => (app.resetDbConfirmOpen = false)}
/>

<ConfirmDialog
  open={app.closeAllConfirmOpen}
  title={t("tab.closeAll")}
  message={t("tab.closeAllConfirm")}
  confirmLabel={t("tab.closeAll")}
  onConfirm={app.confirmCloseAllTabs}
  onCancel={() => (app.closeAllConfirmOpen = false)}
/>

<ConfirmDialog
  open={app.shutdownConfirmOpen}
  title={t("settings.shutdown")}
  message={t("settings.shutdownConfirm")}
  confirmLabel={t("settings.shutdown")}
  onConfirm={app.confirmShutdown}
  onCancel={() => (app.shutdownConfirmOpen = false)}
/>

<ConfirmDialog
  open={app.resetBrowserConfirmOpen}
  title={t("settings.resetBrowser")}
  message={t("settings.resetBrowserConfirm")}
  confirmLabel={t("settings.resetBrowser")}
  onConfirm={app.confirmResetBrowser}
  onCancel={() => (app.resetBrowserConfirmOpen = false)}
/>

<ConfirmDialog
  open={app.restartReticulumConfirmOpen}
  title={t("settings.restartReticulum")}
  message={t("settings.restartReticulumConfirm")}
  confirmLabel={t("settings.restartReticulum")}
  onConfirm={app.confirmRestartReticulum}
  onCancel={() => (app.restartReticulumConfirmOpen = false)}
/>

<ConfirmDialog
  open={app.transportMobileConfirmOpen}
  title={t("settings.enableTransport")}
  message={t("settings.enableTransportMobileConfirm")}
  confirmLabel={t("settings.enableTransport")}
  onConfirm={app.confirmEnableTransportMobile}
  onCancel={() => (app.transportMobileConfirmOpen = false)}
/>

<ConfirmDialog
  open={app.clearHistoryConfirmOpen}
  title={t("history.clear")}
  message={t("history.clearConfirm")}
  confirmLabel={t("history.clear")}
  onConfirm={app.confirmClearHistory}
  onCancel={() => (app.clearHistoryConfirmOpen = false)}
/>

<style>
  .identify-always {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    color: var(--ren-fg-secondary);
    cursor: pointer;
    user-select: none;
  }

  .identify-always input {
    accent-color: var(--ren-accent);
    margin: 0;
  }
</style>
