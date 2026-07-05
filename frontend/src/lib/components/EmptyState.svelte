<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import type { Snippet } from "svelte";

  type Props = {
    title: string;
    description?: string;
    children?: Snippet;
  };

  let { title, description = "", children }: Props = $props();
</script>

<div class="empty-state">
  {#if children}
    <div class="icon" aria-hidden="true">
      {@render children()}
    </div>
  {/if}
  <p class="title">{title}</p>
  {#if description}
    <p class="description">{description}</p>
  {/if}
</div>

<style>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.45rem;
    padding: 2.5rem 1.25rem;
    text-align: center;
    min-height: 8rem;
  }

  .icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 3rem;
    height: 3rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--ren-accent) 12%, var(--ren-chrome-bg));
    color: var(--ren-muted);
    margin-bottom: 0.25rem;
  }

  .title {
    margin: 0;
    color: var(--ren-fg);
    font-size: 0.95rem;
    font-weight: 600;
  }

  .description {
    margin: 0;
    max-width: 22rem;
    color: var(--ren-muted);
    font-size: 0.88rem;
    line-height: 1.45;
  }
</style>
