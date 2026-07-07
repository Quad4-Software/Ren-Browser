<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();
</script>

<ConfirmDialog
  open={app.identifyConfirmOpen}
  title={t("dialog.identifyTitle")}
  message={t("dialog.identifyMessage")}
  confirmLabel={t("common.identify")}
  onConfirm={app.confirmIdentify}
  onCancel={() => (app.identifyConfirmOpen = false)}
/>

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
  open={app.clearHistoryConfirmOpen}
  title={t("history.clear")}
  message={t("history.clearConfirm")}
  confirmLabel={t("history.clear")}
  onConfirm={app.confirmClearHistory}
  onCancel={() => (app.clearHistoryConfirmOpen = false)}
/>
