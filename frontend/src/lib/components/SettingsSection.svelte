<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { ChevronDown } from "@lucide/svelte";
  import type { Snippet } from "svelte";

  type Props = {
    id: string;
    title: string;
    collapsed: boolean;
    onToggle: (id: string) => void;
    children: Snippet;
    heading?: "h2" | "h3";
  };

  let { id, title, collapsed, onToggle, children, heading = "h3" }: Props = $props();
</script>

<section class="settings-section">
  <button
    type="button"
    class="section-toggle"
    aria-expanded={!collapsed}
    aria-controls={`settings-section-${id}`}
    onclick={() => onToggle(id)}
  >
    <span class="chevron" class:collapsed>
      <ChevronDown size={16} />
    </span>
    {#if heading === "h2"}
      <h2>{title}</h2>
    {:else}
      <h3>{title}</h3>
    {/if}
  </button>
  {#if !collapsed}
    <div class="section-body" id={`settings-section-${id}`}>
      {@render children()}
    </div>
  {/if}
</section>

<style>
  .settings-section {
    display: grid;
    gap: 0.65rem;
  }

  .section-toggle {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    width: 100%;
    padding: 0;
    border: none;
    background: transparent;
    color: inherit;
    cursor: pointer;
    text-align: left;
  }

  .section-toggle h2,
  .section-toggle h3 {
    margin: 0;
    flex: 1;
  }

  .section-toggle h2 {
    font-size: 1rem;
    font-weight: 600;
    color: var(--ren-fg);
  }

  .section-toggle h3 {
    margin-top: 0.15rem;
    color: var(--ren-muted);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
  }

  .chevron {
    flex-shrink: 0;
    display: inline-flex;
    color: var(--ren-muted);
    transition: transform 0.15s ease;
  }

  .chevron.collapsed {
    transform: rotate(-90deg);
  }

  .section-body {
    display: grid;
    gap: 0.85rem;
  }
</style>
