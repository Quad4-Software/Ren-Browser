<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Switch } from "bits-ui";

  type Props = {
    checked: boolean;
    label: string;
    onchange: (checked: boolean) => void;
  };

  let { checked, label, onchange }: Props = $props();
</script>

<label class="toggle">
  <span class="label">{label}</span>
  <Switch.Root {checked} onCheckedChange={onchange} aria-label={label} class="ren-switch-track">
    <Switch.Thumb class="ren-switch-thumb" />
  </Switch.Root>
</label>

<style>
  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    min-width: 0;
  }

  .label {
    flex: 1;
    min-width: 0;
    overflow-wrap: break-word;
    word-break: normal;
    color: var(--ren-fg);
    font-size: 0.9rem;
  }

  :global(.ren-switch-track) {
    position: relative;
    flex-shrink: 0;
    width: 2.6rem;
    height: 1.45rem;
    border: 1px solid var(--ren-border);
    border-radius: 999px;
    background: var(--ren-input-bg);
    cursor: pointer;
    padding: 0;
    transition:
      background 0.15s ease,
      border-color 0.15s ease;
  }

  :global(.ren-switch-track[data-state="checked"]) {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
  }

  :global(.ren-switch-thumb) {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 1.05rem;
    height: 1.05rem;
    border-radius: 50%;
    background: var(--ren-fg);
    transition: transform 0.15s ease;
    display: block;
  }

  :global(.ren-switch-thumb[data-state="checked"]) {
    transform: translateX(1.15rem);
    background: #fff;
  }
</style>
