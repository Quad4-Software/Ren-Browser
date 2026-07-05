<script lang="ts">
  type Props = {
    checked: boolean;
    label: string;
    onchange: (checked: boolean) => void;
  };

  let { checked, label, onchange }: Props = $props();
</script>

<label class="toggle">
  <span class="label">{label}</span>
  <button
    type="button"
    role="switch"
    class="track"
    aria-checked={checked}
    aria-label={label}
    onclick={() => onchange(!checked)}
  >
    <span class="thumb" class:on={checked}></span>
  </button>
</label>

<style>
  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .label {
    color: var(--ren-fg);
    font-size: 0.9rem;
  }

  .track {
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

  .track[aria-checked="true"] {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
  }

  .thumb {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 1.05rem;
    height: 1.05rem;
    border-radius: 50%;
    background: var(--ren-fg);
    transition: transform 0.15s ease;
  }

  .thumb.on {
    transform: translateX(1.15rem);
    background: #fff;
  }
</style>
