<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  type Props = {
    message: string;
    isError?: boolean;
  };

  let { message, isError = false }: Props = $props();
</script>

{#if message}
  <div class="plugin-toast" class:error={isError} role="status">{message}</div>
{/if}

<style>
  .plugin-toast {
    position: fixed;
    left: 50%;
    bottom: 1.25rem;
    transform: translateX(-50%);
    z-index: 60;
    max-width: min(32rem, calc(100% - 2rem));
    padding: 0.55rem 0.9rem;
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    border: 1px solid var(--ren-border);
    box-shadow: var(--ren-shadow);
    font-size: 0.85rem;
    user-select: text;
    cursor: text;
    word-break: break-word;
  }

  .plugin-toast.error {
    border-color: color-mix(in srgb, #ef4444 55%, var(--ren-border));
    background: color-mix(in srgb, #ef4444 12%, var(--ren-chrome-bg));
    color: #ef4444;
  }
</style>
