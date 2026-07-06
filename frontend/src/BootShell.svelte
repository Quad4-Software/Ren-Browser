<script lang="ts">
  import type { Component } from "svelte";

  type Props = {
    Root: Component;
  };

  let { Root }: Props = $props();
  let fault = $state<string | null>(null);

  function onerror(error: unknown) {
    fault = error instanceof Error ? error.message : "Unexpected render error";
  }

  function retry() {
    fault = null;
    location.reload();
  }
</script>

<svelte:boundary {onerror}>
  <Root />

  {#snippet failed()}
    <div class="boot-fault">
      <h1>Ren Browser hit a rendering error</h1>
      {#if fault}
        <p>{fault}</p>
      {/if}
      <button type="button" onclick={retry}>Reload</button>
    </div>
  {/snippet}
</svelte:boundary>

<style>
  .boot-fault {
    padding: 24px;
    color: #f3f4f6;
    font-family: system-ui, sans-serif;
    line-height: 1.5;
  }

  .boot-fault h1 {
    margin: 0 0 12px;
    font-size: 1.1rem;
  }

  .boot-fault p {
    margin: 0 0 16px;
    color: #9ca3af;
  }

  .boot-fault button {
    border: 1px solid #3f3f46;
    background: #18181b;
    color: #f3f4f6;
    border-radius: 8px;
    padding: 8px 14px;
    cursor: pointer;
  }
</style>
