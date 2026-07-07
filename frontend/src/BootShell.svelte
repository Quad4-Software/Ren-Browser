<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import type { Component } from "svelte";
  import CrashPage from "$lib/components/CrashPage.svelte";
  import { crashErrorMessage } from "$lib/browser/crash-log";
  import { Shutdown } from "../bindings/renbrowser/internal/app/browserservice";

  type Props = {
    Root: Component;
  };

  let { Root }: Props = $props();
  let fault = $state<string | null>(null);
  let faultCause = $state<unknown>(null);
  let closing = $state(false);

  function onerror(error: unknown) {
    faultCause = error;
    fault = crashErrorMessage(error);
  }

  function retry() {
    fault = null;
    faultCause = null;
    location.reload();
  }

  async function closeApp() {
    if (closing) {
      return;
    }
    closing = true;
    try {
      await Shutdown();
    } catch {
      window.close();
    }
  }
</script>

<svelte:boundary {onerror}>
  <Root />

  {#snippet failed()}
    <CrashPage
      message={fault ?? ""}
      cause={faultCause}
      {closing}
      onReload={retry}
      onClose={() => void closeApp()}
    />
  {/snippet}
</svelte:boundary>
